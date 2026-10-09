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
	mrand "math/rand"
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

const UserAgent = "DirRunner v1.3.0"

// randomUserAgents is a small pool of common, real-world User-Agent strings
// rotated per request when the --random-agent flag is set, to avoid trivial
// User-Agent-based filtering.
var randomUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64; rv:126.0) Gecko/20100101 Firefox/126.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
}

// RandomUserAgent returns a User-Agent chosen at random from the pool.
func RandomUserAgent() string {
	return randomUserAgents[mrand.Intn(len(randomUserAgents))]
}

type HTTPOptions struct {
	Method          string
	UserAgent       string
	RandomUserAgent bool
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
	ua := opts.UserAgent
	switch {
	case opts.RandomUserAgent:
		ua = RandomUserAgent()
	case ua == "":
		ua = UserAgent
	}
	req.Header.Set("User-Agent", ua)
	if opts.Cookie != "" {
		req.Header.Set("Cookie", opts.Cookie)
	}
	for k, v := range opts.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		// Go's HTTP client ignores a "Host" entry in Header; the Host header
		// must be set through req.Host instead. This is what makes vhost
		// enumeration (and FUZZ against the Host header) actually work.
		if strings.EqualFold(k, "Host") {
			req.Host = v
			continue
		}
		req.Header.Set(k, v)
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

// maxDrainBytes caps how much of a response body we read just to enable
// keep-alive connection reuse. Small bodies are drained so the connection is
// pooled; large bodies (e.g. a site that serves its full HTML homepage on every
// 404) are left unread so we never waste bandwidth downloading content a
// directory scanner does not need.
const maxDrainBytes = 64 * 1024

func ResponseSize(resp *http.Response) int64 {
	if resp.ContentLength >= 0 {
		// Size is already known from the header. Only drain small bodies so the
		// connection can be reused; skip large ones to keep requests fast.
		if resp.ContentLength <= maxDrainBytes {
			io.Copy(io.Discard, resp.Body)
		}
		return resp.ContentLength
	}
	// Unknown length (e.g. chunked): we must read the body to measure it.
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
