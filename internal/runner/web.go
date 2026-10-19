package runner

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"dirrunner/internal/output"
)

// maxBodyBytes caps how much of a response body the web modules (traversal,
// bypass, sqli, crawl, clone) read into memory. These modules, unlike the
// directory scanner, need the actual body to compare lengths, search for
// patterns or extract links, so the body is read in full up to this limit.
const maxBodyBytes = 10 << 20 // 10 MiB

// FetchResponse performs a single request and returns the status code, response
// headers and the full body (capped at maxBodyBytes). extra headers are merged
// on top of opts.Headers for this request only, which is how the bypass and
// sqli modules inject per-request headers without mutating shared options.
func FetchResponse(ctx context.Context, client *http.Client, opts HTTPOptions, method, rawURL string, extra map[string]string) (int, http.Header, []byte, error) {
	o := opts
	if method != "" {
		o.Method = method
	}
	if len(extra) > 0 {
		merged := make(map[string]string, len(opts.Headers)+len(extra))
		for k, v := range opts.Headers {
			merged[k] = v
		}
		for k, v := range extra {
			merged[k] = v
		}
		o.Headers = merged
	}
	req, err := NewRequest(ctx, o.Method, rawURL, nil, o)
	if err != nil {
		return 0, nil, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	return resp.StatusCode, resp.Header, body, err
}

// detectReadCap bounds how much of a response body the detection modules
// (traversal, bypass, sqli) read. These modules only need the body's length and
// a prefix to match signatures or error strings, not the whole document, so
// reading a cap instead of up to maxBodyBytes keeps thousands of probe requests
// cheap. Bodies at or below the cap are read to EOF, which also lets the
// connection be reused; larger bodies are left unread (dropped) rather than
// downloaded in full.
const detectReadCap = 256 << 10 // 256 KiB

// fetchDetectBody performs a request (with an optional request body) for the
// detection modules and returns the status, a capped body prefix (for
// signature/error matching), the response size, and the wall-clock time. The
// size is taken from Content-Length when the server provides it, so length
// comparisons stay accurate even though only a prefix is read.
func fetchDetectBody(ctx context.Context, client *http.Client, opts HTTPOptions, method, rawURL, reqBody string, extra map[string]string) (int, []byte, int64, time.Duration, error) {
	o := opts
	if method != "" {
		o.Method = method
	}
	if len(extra) > 0 {
		merged := make(map[string]string, len(opts.Headers)+len(extra))
		for k, v := range opts.Headers {
			merged[k] = v
		}
		for k, v := range extra {
			merged[k] = v
		}
		o.Headers = merged
	}
	var body io.Reader
	if reqBody != "" {
		body = strings.NewReader(reqBody)
	}
	req, err := NewRequest(ctx, o.Method, rawURL, body, o)
	if err != nil {
		return 0, nil, 0, 0, err
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, 0, time.Since(start), err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, detectReadCap))
	size := int64(len(respBody))
	if resp.ContentLength >= 0 {
		size = resp.ContentLength
	}
	return resp.StatusCode, respBody, size, time.Since(start), nil
}

// fetchDetect is fetchDetectBody without a request body or timing.
func fetchDetect(ctx context.Context, client *http.Client, opts HTTPOptions, method, rawURL string, extra map[string]string) (int, []byte, int64, error) {
	status, body, size, _, err := fetchDetectBody(ctx, client, opts, method, rawURL, "", extra)
	return status, body, size, err
}

// fetchDetectTimed is fetchDetectBody without a request body.
func fetchDetectTimed(ctx context.Context, client *http.Client, opts HTTPOptions, method, rawURL string, extra map[string]string) (int, []byte, int64, time.Duration, error) {
	return fetchDetectBody(ctx, client, opts, method, rawURL, "", extra)
}

// ValidURL reports whether raw is an absolute http(s) URL with a host, matching
// WebRunner's formats.validate_url behaviour.
func ValidURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// RunPool runs fn over items with a bounded worker pool, streaming each result
// live and collecting them for the final (JSON) output. It mirrors
// RunWorkersMany but takes an in-memory slice of jobs instead of a wordlist
// file, so the generated-payload modules can reuse the same concurrency and
// progress plumbing.
func RunPool[T any](ctx context.Context, label string, items []T, workers int, fn func(context.Context, T) []output.Result) ([]output.Result, error) {
	if workers < 1 {
		workers = 1
	}
	total := len(items)
	output.Info("starting %s with %d workers and %d tasks", label, workers, total)
	progress := output.NewProgress("Progress", total)
	defer progress.Finish()

	jobs := make(chan T, workers*2)
	results := make(chan output.Result, workers)

	go func() {
		defer close(jobs)
		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case jobs <- item:
			}
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
				case job, ok := <-jobs:
					if !ok {
						return
					}
					for _, result := range fn(ctx, job) {
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
	}

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return found, err
	}
	return found, nil
}

// HTML extraction patterns. These are deliberately regex-based rather than a
// full HTML parser so the project stays dependency-free; they are good enough
// for recon-grade link and asset discovery.
var (
	reHrefAttr  = regexp.MustCompile(`(?i)<a\b[^>]*?\bhref\s*=\s*["']?([^"'\s>]+)`)
	reScriptSrc = regexp.MustCompile(`(?i)<script\b[^>]*?\bsrc\s*=\s*["']?([^"'\s>]+)`)
	reImgSrc    = regexp.MustCompile(`(?i)<img\b[^>]*?\bsrc\s*=\s*["']?([^"'\s>]+)`)
	reLinkHref  = regexp.MustCompile(`(?i)<link\b[^>]*?\bhref\s*=\s*["']?([^"'\s>]+)`)
	// reEmail requires a letters-only TLD and word boundaries so it does not
	// swallow a trailing dot ("a@b.io.") or a word glued to the address
	// ("a@b.comVERSION"); the domain is one or more dot-separated labels.
	reEmail = regexp.MustCompile(`(?i)\b[a-z0-9][a-z0-9._%+\-]*@(?:[a-z0-9](?:[a-z0-9\-]*[a-z0-9])?\.)+[a-z]{2,24}\b`)
)

// emailExtBlocklist holds "TLDs" that are really asset extensions, so retina
// images and similar (logo@2x.png, icon@3x.svg) are not reported as e-mails.
var emailExtBlocklist = map[string]struct{}{
	"png": {}, "jpg": {}, "jpeg": {}, "gif": {}, "webp": {}, "svg": {},
	"ico": {}, "bmp": {}, "css": {}, "js": {}, "mjs": {}, "map": {},
	"woff": {}, "woff2": {}, "ttf": {}, "eot": {}, "otf": {}, "mp4": {},
	"webm": {}, "mp3": {}, "pdf": {}, "zip": {}, "min": {},
}

// extractMatches returns the first capture group of every match of re in body.
func extractMatches(re *regexp.Regexp, body string) []string {
	matches := re.FindAllStringSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, m[1])
		}
	}
	return out
}

// resolveURL turns a possibly-relative reference into an absolute URL against
// base, dropping fragments. It returns "" when the reference cannot be used.
func resolveURL(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "#") ||
		strings.HasPrefix(strings.ToLower(ref), "javascript:") ||
		strings.HasPrefix(strings.ToLower(ref), "mailto:") ||
		strings.HasPrefix(strings.ToLower(ref), "data:") {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	resolved := b.ResolveReference(r)
	resolved.Fragment = ""
	return resolved.String()
}

// applyEncodings replaces each base payload with its chain-encoded form. When
// --encode lists several encodings they are applied as a pipeline in the given
// order and only the final result is sent (the plain payload is not). For
// example --encode xor,base64,urlencode sends urlencode(base64(xor(payload))).
// With no encodings the plain payloads are returned unchanged. Shared by the
// traversal and sqli modules.
func applyEncodings(base, encodings []string, xorKey string) []string {
	if len(encodings) == 0 {
		return base
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(base))
	for _, payload := range base {
		encoded := chainEncode(payload, encodings, xorKey)
		if _, ok := seen[encoded]; ok {
			continue
		}
		seen[encoded] = struct{}{}
		out = append(out, encoded)
	}
	return out
}

// chainEncode applies the encodings to payload as a pipeline, left to right, and
// returns the final, URL-transmittable string. base64 and urlencode already emit
// ASCII; a trailing xor step leaves raw bytes, so it is percent-encoded so the
// value can travel in a URL.
func chainEncode(payload string, encodings []string, xorKey string) string {
	if xorKey == "" {
		xorKey = "DirRunner"
	}
	data := []byte(payload)
	last := ""
	for _, enc := range encodings {
		switch enc {
		case "base64":
			data = []byte(base64.StdEncoding.EncodeToString(data))
		case "urlencode":
			data = []byte(strings.ReplaceAll(url.QueryEscape(string(data)), "+", "%20"))
		case "xor":
			data = xorBytes(data, []byte(xorKey))
		default:
			continue
		}
		last = enc
	}
	if last == "xor" {
		return percentEncodeAll(data)
	}
	return string(data)
}

// xorBytes XORs raw against a repeating key.
func xorBytes(raw, key []byte) []byte {
	if len(key) == 0 {
		key = []byte("DirRunner")
	}
	out := make([]byte, len(raw))
	for i, b := range raw {
		out[i] = b ^ key[i%len(key)]
	}
	return out
}

// dedupeEmails returns the unique, plausible e-mails found in body, preserving
// order. Matches whose TLD is really an asset extension (e.g. logo@2x.png) are
// dropped, and comparison is case-insensitive so Foo@X.com and foo@x.com count
// once.
func dedupeEmails(body string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, email := range reEmail.FindAllString(body, -1) {
		if dot := strings.LastIndexByte(email, '.'); dot >= 0 {
			if _, bad := emailExtBlocklist[strings.ToLower(email[dot+1:])]; bad {
				continue
			}
		}
		key := strings.ToLower(email)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, email)
	}
	return out
}
