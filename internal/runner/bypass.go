package runner

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"dirrunner/internal/output"
)

// Default wordlists for the bypass module, editable under wordlist/ and
// overridable per run.
const (
	DefaultBypassHeaders = "wordlist/bypass-headers.txt"
	DefaultBypassPaths   = "wordlist/bypass-paths.txt"
)

type BypassOptions struct {
	Targets           []string
	HeaderWordlist    string   // base "Header-Name: value" list
	PathWordlist      string   // base path-suffix list
	CustomHeaders     []string // extra inline "Header-Name: value" headers
	CustomPathSource  string   // extra file, directory, or single path to test
	ContinueOnSuccess bool
	Workers           int
	HTTP              HTTPOptions
}

// RunBypass tests access-control bypass candidates (header tricks and path
// mutations) against each target and reports any that change the baseline
// 401/403 response.
func RunBypass(ctx context.Context, opts BypassOptions) ([]output.Result, error) {
	baseHeaderLines, err := LoadWordlist(firstNonEmpty(opts.HeaderWordlist, DefaultBypassHeaders))
	if err != nil {
		return nil, fmt.Errorf("loading bypass headers: %w", err)
	}
	bypassHeaders, err := parseBypassHeaders(baseHeaderLines)
	if err != nil {
		return nil, err
	}
	bypassPathSuffixes, err := LoadWordlist(firstNonEmpty(opts.PathWordlist, DefaultBypassPaths))
	if err != nil {
		return nil, fmt.Errorf("loading bypass paths: %w", err)
	}
	customHeaders, err := parseBypassHeaders(opts.CustomHeaders)
	if err != nil {
		return nil, err
	}
	customPaths, err := loadBypassPaths(opts.CustomPathSource)
	if err != nil {
		return nil, err
	}

	client := NewHTTPClient(opts.HTTP)
	var all []output.Result
	for _, target := range opts.Targets {
		if ctx.Err() != nil {
			break
		}
		output.Info("checking access control for %s", target)
		baseStatus, _, _, err := fetchDetect(ctx, client, opts.HTTP, "GET", target, nil)
		if err != nil {
			output.Warn("baseline request failed for %s: %v", target, err)
			continue
		}
		output.Info("baseline status for %s: %d", target, baseStatus)
		if baseStatus == 401 || baseStatus == 403 {
			output.Warn("%s returned %d; testing bypass candidates", target, baseStatus)
		} else {
			output.Info("%s did not return 401/403, running light bypass probes anyway", target)
		}

		candidates := buildBypassCandidates(target, bypassHeaders, customHeaders, bypassPathSuffixes, customPaths)
		poolCtx, cancel := context.WithCancel(ctx)
		var found atomic.Bool
		results, err := RunPool(poolCtx, "bypass", candidates, opts.Workers, func(ctx context.Context, c bypassCandidate) []output.Result {
			if !opts.ContinueOnSuccess && found.Load() {
				return nil
			}
			output.Debug("testing %s %s", c.kind, c.url)
			status, _, size, err := fetchDetect(ctx, client, opts.HTTP, "GET", c.url, c.headers)
			if err != nil {
				return nil
			}
			if !bypassHit(status, baseStatus) {
				return nil
			}
			found.Store(true)
			if !opts.ContinueOnSuccess {
				cancel()
			}
			detail := c.kind
			if len(c.headers) > 0 {
				detail = fmt.Sprintf("%s %s", c.kind, headerString(c.headers))
			}
			output.Warn("%s reached %d (baseline %d) at %s", detail, status, baseStatus, c.url)
			return []output.Result{{
				Type:   "bypass",
				Target: target,
				Method: "GET",
				URL:    c.url,
				Status: status,
				Size:   size,
				Value:  detail,
			}}
		})
		cancel()
		if err != nil {
			return all, err
		}
		all = append(all, results...)
	}
	return all, nil
}

// bypassCandidate is one access-control bypass probe: either a header set sent
// to the original URL, or a mutated URL requested plainly.
type bypassCandidate struct {
	kind    string
	url     string
	headers map[string]string
}

// urlOverrideHeaders route the request to the path in their value. The bypass is
// to request an allowed path while pointing one of these at the protected path.
var urlOverrideHeaders = []string{
	"X-Original-URL",
	"X-Rewrite-URL",
	"X-Override-URL",
	"X-HTTP-DestinationURL",
	"X-Original-Url",
}

// buildBypassCandidates flattens header, path-suffix and custom-path probes into
// a single list so they can be tested concurrently.
func buildBypassCandidates(target string, baseHeaders, customHeaders []map[string]string, suffixes, customPaths []string) []bypassCandidate {
	var candidates []bypassCandidate
	for _, h := range baseHeaders {
		candidates = append(candidates, bypassCandidate{kind: "header", url: target, headers: h})
	}
	for _, h := range customHeaders {
		candidates = append(candidates, bypassCandidate{kind: "header", url: target, headers: h})
	}
	for _, payload := range suffixes {
		for _, v := range bypassPathVariants(target, payload) {
			candidates = append(candidates, bypassCandidate{kind: "path", url: v})
		}
	}
	for _, candidate := range customPaths {
		candidates = append(candidates, bypassCandidate{kind: "custom-path", url: bypassCustomPathVariant(target, candidate)})
	}
	// URL-override headers: request a dummy path that normally 404s, pointing the
	// override header at the protected path. A 200 means the override served the
	// protected content (the "URL-based access control circumvented" bypass).
	if u, err := url.Parse(target); err == nil && u.Path != "" && u.Path != "/" {
		dummy := u.Scheme + "://" + u.Host + "/" + RandomToken()[:8]
		for _, h := range urlOverrideHeaders {
			candidates = append(candidates, bypassCandidate{kind: "url-override " + h, url: dummy, headers: map[string]string{h: u.Path}})
		}
	}
	return candidates
}

func parseBypassHeaders(raw []string) ([]map[string]string, error) {
	var headers []map[string]string
	for _, h := range raw {
		name, value, ok := strings.Cut(h, ":")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, fmt.Errorf("invalid --bypass-header value %q, expected 'Header-Name: value'", h)
		}
		headers = append(headers, map[string]string{name: strings.TrimSpace(value)})
	}
	return headers, nil
}

func loadBypassPaths(source string) ([]string, error) {
	if source == "" {
		return nil, nil
	}
	info, err := os.Stat(source)
	if err != nil {
		// Not a filesystem path: treat the value itself as a single candidate.
		return []string{source}, nil
	}
	if info.IsDir() {
		var paths []string
		err := filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(source, path)
			if err != nil || rel == "." {
				return nil
			}
			rel = filepath.ToSlash(rel)
			entry := "/" + strings.TrimRight(rel, "/")
			if d.IsDir() {
				entry += "/"
			}
			paths = append(paths, entry)
			return nil
		})
		return paths, err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

// bypassPathVariants returns the URL mutations for one path payload. An
// access-control front-end often matches on the literal request path while the
// back-end normalises it differently, so each payload is tried at the positions
// that exploit that gap: as a suffix after the protected segment
// (`/admin/..;/`), as a prefix before it (`/..;/admin`), and — for up-traversal
// tokens like `..;/` — wrapped so the path walks up and back into the segment
// (`/admin/..;/admin`). Appending only as a suffix would turn an up-traversal
// payload into a request for the parent directory, which never bypasses the
// protection on the segment itself.
func bypassPathVariants(target, payload string) []string {
	u, err := url.Parse(target)
	if err != nil {
		return []string{target}
	}
	base := u.Scheme + "://" + u.Host
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	// Query/fragment tricks only make sense appended to the whole path.
	switch payload {
	case "?", "/#", "/?foo=bar", "/;foo=bar/":
		return []string{base + strings.TrimRight(path, "/") + payload}
	}
	trimmed := strings.TrimRight(path, "/")
	dir, seg := splitLastSegment(trimmed)
	variants := []string{base + trimmed + payload} // suffix
	if seg != "" {
		variants = append(variants, base+dir+payload+seg) // prefix
		low := strings.ToLower(payload)
		if strings.Contains(low, "..") || strings.Contains(low, "%2e%2e") || strings.Contains(low, "%252e") {
			variants = append(variants, base+trimmed+payload+seg) // wrap
		}
	}
	return dedupeStrings(variants)
}

// splitLastSegment splits a path into its parent directory and final segment,
// e.g. "/api/admin" -> ("/api", "admin") and "/admin" -> ("", "admin").
func splitLastSegment(p string) (dir, seg string) {
	p = strings.TrimRight(p, "/")
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return "", p
	}
	return p[:i], p[i+1:]
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := in[:0]
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// bypassPathVariant appends suffix to the target's path. The URL is assembled
// by concatenation so encoded suffixes (e.g. "/%2e/") reach the server as
// written instead of being re-escaped.
func bypassPathVariant(target, suffix string) string {
	u, err := url.Parse(target)
	if err != nil {
		return target
	}
	base := u.Scheme + "://" + u.Host
	path := u.EscapedPath()
	if suffix == "?" {
		if path == "" {
			path = "/"
		}
		return base + path
	}
	if path == "" {
		path = "/"
	}
	return base + strings.TrimRight(path, "/") + suffix
}

func bypassCustomPathVariant(target, candidate string) string {
	if strings.HasPrefix(candidate, "http://") || strings.HasPrefix(candidate, "https://") {
		return candidate
	}
	u, err := url.Parse(target)
	if err != nil {
		return target
	}
	base := u.Scheme + "://" + u.Host
	if strings.HasPrefix(candidate, "/") {
		return base + candidate
	}
	return base + "/" + strings.TrimLeft(candidate, "/")
}

// bypassHit reports whether a candidate response is a genuine bypass: a 2xx
// success that was not already reachable at baseline. The previous
// "status < 400 || status != baseStatus" rule flagged any change (e.g. a 403
// becoming a 404 or 500), which produced noise rather than bypasses.
func bypassHit(status, baseStatus int) bool {
	success := status >= 200 && status < 300
	baseReachable := baseStatus >= 200 && baseStatus < 300
	return success && !baseReachable
}

func headerString(headers map[string]string) string {
	parts := make([]string, 0, len(headers))
	for k, v := range headers {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
