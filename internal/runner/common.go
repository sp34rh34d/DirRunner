package runner

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"dirrunner/internal/output"
)

const UserAgent = "DirRunner v2.0"

type HTTPOptions struct {
	Method          string
	UserAgent       string
	Cookie          string
	Username        string
	Password        string
	Timeout         time.Duration
	TLSVerify       bool
	FollowRedirects bool
	Headers         map[string]string
	ExcludeSizes    map[int64]struct{}
	ExcludeRanges   []SizeRange
	Tor             bool
	TorProxy        string
}

type SizeRange struct {
	Min int64
	Max int64
}

func NewHTTPClient(opts HTTPOptions) *http.Client {
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: !opts.TLSVerify},
		MaxIdleConns:        256,
		MaxIdleConnsPerHost: 256,
		IdleConnTimeout:     30 * time.Second,
	}
	if opts.Tor {
		proxyAddr := opts.TorProxy
		if proxyAddr == "" {
			proxyAddr = "127.0.0.1:9050"
		}
		transport.DialContext = socks5DialContext(proxyAddr)
		output.Debug("using Tor SOCKS5 proxy %s", proxyAddr)
	}
	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
	}
	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client
}

func socks5DialContext(proxyAddr string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		var dialer net.Dialer
		conn, err := dialer.DialContext(ctx, network, proxyAddr)
		if err != nil {
			return nil, err
		}
		if err := socks5Connect(ctx, conn, address); err != nil {
			conn.Close()
			return nil, err
		}
		return conn, nil
	}
}

func socks5Connect(ctx context.Context, conn net.Conn, address string) error {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		return err
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if reply[0] != 0x05 || reply[1] != 0x00 {
		return fmt.Errorf("socks5 proxy does not allow no-auth connection")
	}

	host, portRaw, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid target port %q", portRaw)
	}

	req := []byte{0x05, 0x01, 0x00}
	if ip := net.ParseIP(host); ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			req = append(req, 0x01)
			req = append(req, ipv4...)
		} else {
			req = append(req, 0x04)
			req = append(req, ip.To16()...)
		}
	} else {
		if len(host) > 255 {
			return fmt.Errorf("target host too long for socks5: %s", host)
		}
		req = append(req, 0x03, byte(len(host)))
		req = append(req, []byte(host)...)
	}
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	req = append(req, portBytes...)
	if _, err := conn.Write(req); err != nil {
		return err
	}

	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != 0x05 || header[1] != 0x00 {
		return fmt.Errorf("socks5 connect failed with code 0x%02x", header[1])
	}
	var discard int
	switch header[3] {
	case 0x01:
		discard = 4
	case 0x03:
		lenByte := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenByte); err != nil {
			return err
		}
		discard = int(lenByte[0])
	case 0x04:
		discard = 16
	default:
		return fmt.Errorf("invalid socks5 address type 0x%02x", header[3])
	}
	if discard > 0 {
		if _, err := io.CopyN(io.Discard, conn, int64(discard)); err != nil {
			return err
		}
	}
	if _, err := io.CopyN(io.Discard, conn, 2); err != nil {
		return err
	}
	return nil
}

func NewRequest(ctx context.Context, method, rawURL string, body io.Reader, opts HTTPOptions) (*http.Request, error) {
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), rawURL, body)
	if err != nil {
		return nil, err
	}
	if opts.UserAgent == "" {
		opts.UserAgent = UserAgent
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	if opts.Cookie != "" {
		req.Header.Set("Cookie", opts.Cookie)
	}
	for k, v := range opts.Headers {
		if strings.TrimSpace(k) != "" {
			req.Header.Set(k, v)
		}
	}
	if opts.Username != "" || opts.Password != "" {
		req.SetBasicAuth(opts.Username, opts.Password)
	}
	return req, nil
}

func RunWorkers(ctx context.Context, wordlist string, workers int, fn func(context.Context, string) (output.Result, bool)) ([]output.Result, error) {
	return RunWorkersMany(ctx, wordlist, workers, func(ctx context.Context, word string) []output.Result {
		if result, found := fn(ctx, word); found {
			return []output.Result{result}
		}
		return nil
	})
}

func RunWorkersMany(ctx context.Context, wordlist string, workers int, fn func(context.Context, string) []output.Result) ([]output.Result, error) {
	if workers < 1 {
		workers = 1
	}
	output.Info("loading wordlist %s...", wordlist)
	words, err := LoadWordlist(wordlist)
	if err != nil {
		return nil, err
	}
	total := len(words)
	output.Info("starting enumeration with %d workers and %d words", workers, total)
	output.Debug("starting %d workers with %d words from %s", workers, total, wordlist)
	progress := output.NewProgress("Progress", total)
	defer progress.Finish()
	jobs := make(chan string, workers*2)
	results := make(chan output.Result, workers)

	go func() {
		defer close(jobs)
		for _, word := range words {
			jobs <- word
		}
	}()

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case word, ok := <-jobs:
					if !ok {
						return
					}
					for _, result := range fn(ctx, word) {
						results <- result
					}
					progress.Advance()
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var found []output.Result
	for result := range results {
		found = append(found, result)
		output.PrintLiveResult(result)
		if result.Type != "dns" {
			output.Debug("found %s %s", result.Type, resultDisplay(result))
		}
	}

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return found, err
	}
	return found, nil
}

func LoadWordlist(path string) ([]string, error) {
	start := time.Now()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	words := parseWordlist(string(data))
	output.Debug("loaded %d words from %s in %s", len(words), path, time.Since(start).Round(time.Millisecond))
	return words, nil
}

func parseWordlist(raw string) []string {
	lines := strings.Split(raw, "\n")
	words := make([]string, 0, len(lines))
	for _, line := range lines {
		word := strings.TrimSpace(line)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		words = append(words, word)
	}
	return words
}

func resultDisplay(result output.Result) string {
	if result.URL != "" {
		return result.URL
	}
	if result.Host != "" {
		return result.Host
	}
	return result.Target
}

func StatusSet(codes []int) map[int]struct{} {
	set := make(map[int]struct{}, len(codes))
	for _, code := range codes {
		if code >= 100 && code <= 599 {
			set[code] = struct{}{}
		}
	}
	return set
}

func JoinURL(base, part string, trailingSlash bool) string {
	u, err := url.Parse(base)
	if err != nil {
		return strings.TrimRight(base, "/") + "/" + strings.Trim(part, "/")
	}
	basePath := strings.TrimRight(u.Path, "/")
	u.Path = basePath + "/" + strings.Trim(part, "/")
	if trailingSlash && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	return u.String()
}

func RandomToken() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "dirrunner-missing-check"
	}
	return hex.EncodeToString(b)
}

func ResponseSize(resp *http.Response) int64 {
	if resp.ContentLength >= 0 {
		return resp.ContentLength
	}
	n, _ := io.Copy(io.Discard, resp.Body)
	return n
}

func ExcludedSize(size int64, opts HTTPOptions) bool {
	if _, ok := opts.ExcludeSizes[size]; ok {
		return true
	}
	for _, r := range opts.ExcludeRanges {
		if size >= r.Min && size <= r.Max {
			return true
		}
	}
	return false
}
