package runner

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"io"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"dirrunner/internal/output"
)

const UserAgent = "DirRunner v2.0.0"

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
	Proxy           string // http://, https://, socks5:// or socks5h:// proxy URL
}

type SizeRange struct {
	Min int64
	Max int64
}

func NewHTTPClient(opts HTTPOptions) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !opts.TLSVerify,
			// Cache TLS sessions so the handshakes we cannot avoid (large bodies
			// that drop keep-alive, new hosts) resume instead of doing a full
			// handshake each time.
			ClientSessionCache: tls.NewLRUClientSessionCache(512),
		},
		// HTTP/2 is intentionally left off (setting TLSClientConfig disables Go's
		// automatic h2 upgrade, and we do not force it): forcing h2 breaks against
		// some servers and proxies with "received GOAWAY before SETTINGS" errors.
		// HTTP/1.1 keep-alive plus the TLS session cache already reuse connections.
		MaxIdleConns:        512,
		MaxIdleConnsPerHost: 512,
		IdleConnTimeout:     30 * time.Second,
	}
	if opts.Proxy != "" {
		// Go's transport handles http, https, socks5 and socks5h proxy URLs
		// natively, so a single --proxy covers plain HTTP proxies (Burp, ZAP)
		// and SOCKS5 (Tor: socks5h://127.0.0.1:9050).
		if proxyURL, err := url.Parse(opts.Proxy); err == nil && proxyURL.Host != "" {
			transport.Proxy = http.ProxyURL(proxyURL)
			output.Debug("using proxy %s", opts.Proxy)
		} else {
			output.Warn("invalid --proxy %q, ignoring", opts.Proxy)
		}
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

// loadRawLines reads a wordlist preserving each payload's leading/trailing
// spaces (LoadWordlist trims them, which would corrupt payloads like " AND 1=1").
// It still skips blank lines and # comments.
func loadRawLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, nil
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
