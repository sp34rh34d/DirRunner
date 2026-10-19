package runner

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sync"

	"dirrunner/internal/output"
)

type CrawlOptions struct {
	Targets  []string
	MaxDepth int
	Emails   bool
	Regex    string
	SameHost bool
	Workers  int
	HTTP     HTTPOptions
}

// RunCrawl walks each target breadth-first up to MaxDepth, emitting discovered
// page URLs plus JavaScript and image assets. With Emails it also extracts
// e-mail addresses from every page, and with Regex it reports every match of the
// pattern. It unifies WebRunner's scraping, email-extractor and regx modules.
func RunCrawl(ctx context.Context, opts CrawlOptions) ([]output.Result, error) {
	var re *regexp.Regexp
	if opts.Regex != "" {
		compiled, err := regexp.Compile(opts.Regex)
		if err != nil {
			return nil, fmt.Errorf("invalid --regx pattern: %w", err)
		}
		re = compiled
	}
	client := NewHTTPClient(opts.HTTP)

	// In extraction mode (--emails / --regx) the crawl is a hunt for specific
	// findings, so only emit those; the page/JS/image inventory is suppressed
	// (the crawler still follows links internally to reach more pages).
	extractionMode := opts.Emails || re != nil

	// Shared state across depth levels.
	var mu sync.Mutex
	visited := map[string]struct{}{} // pages already fetched
	emitted := map[string]struct{}{} // type|value already reported (assets, emails, matches)
	var all []output.Result

	// seedHosts records the allowed hosts when SameHost scoping is on.
	seedHosts := map[string]struct{}{}
	for _, t := range opts.Targets {
		if u, err := url.Parse(t); err == nil {
			seedHosts[u.Host] = struct{}{}
		}
	}
	inScope := func(raw string) bool {
		if !opts.SameHost {
			return true
		}
		u, err := url.Parse(raw)
		if err != nil {
			return false
		}
		_, ok := seedHosts[u.Host]
		return ok
	}

	// firstSeen reports whether key has not been emitted yet and records it.
	firstSeen := func(key string) bool {
		mu.Lock()
		defer mu.Unlock()
		if _, ok := emitted[key]; ok {
			return false
		}
		emitted[key] = struct{}{}
		return true
	}

	frontier := append([]string{}, opts.Targets...)
	for depth := 0; depth <= opts.MaxDepth && len(frontier) > 0; depth++ {
		if ctx.Err() != nil {
			break
		}
		// Reduce the frontier to not-yet-visited, in-scope, valid pages.
		var current []string
		for _, raw := range frontier {
			if !ValidURL(raw) {
				continue
			}
			if _, done := visited[raw]; done {
				continue
			}
			visited[raw] = struct{}{}
			current = append(current, raw)
		}
		if len(current) == 0 {
			break
		}

		var next []string
		var nextMu sync.Mutex

		levelResults, err := RunPool(ctx, fmt.Sprintf("crawl depth %d", depth), current, opts.Workers, func(ctx context.Context, pageURL string) []output.Result {
			status, _, body, err := FetchResponse(ctx, client, opts.HTTP, "GET", pageURL, nil)
			if err != nil {
				output.Debug("error fetching %s: %v", pageURL, err)
				return nil
			}
			text := string(body)
			var out []output.Result
			if !extractionMode {
				out = append(out, output.Result{
					Type:   "url",
					Target: pageURL,
					Method: "GET",
					URL:    pageURL,
					Status: status,
					Size:   int64(len(body)),
				})
			}

			if opts.Emails {
				for _, email := range dedupeEmails(text) {
					if firstSeen("email|" + email) {
						out = append(out, output.Result{Type: "email", Target: pageURL, URL: pageURL, Value: email})
					}
				}
			}
			if re != nil {
				for _, match := range re.FindAllString(text, -1) {
					if firstSeen("match|" + match) {
						out = append(out, output.Result{Type: "match", Target: pageURL, URL: pageURL, Value: match})
					}
				}
			}

			// Page links feed the next depth level.
			for _, href := range extractMatches(reHrefAttr, text) {
				if abs := resolveURL(pageURL, href); abs != "" && inScope(abs) {
					nextMu.Lock()
					next = append(next, abs)
					nextMu.Unlock()
				}
			}
			// JS and image assets are part of the discovery inventory, not an
			// extraction target, so they are reported only outside extraction mode.
			if !extractionMode {
				for _, src := range extractMatches(reScriptSrc, text) {
					if abs := resolveURL(pageURL, src); abs != "" && firstSeen("js|"+abs) {
						out = append(out, output.Result{Type: "js", Target: pageURL, URL: abs})
					}
				}
				for _, src := range extractMatches(reImgSrc, text) {
					if abs := resolveURL(pageURL, src); abs != "" && firstSeen("img|"+abs) {
						out = append(out, output.Result{Type: "img", Target: pageURL, URL: abs})
					}
				}
			}
			return out
		})
		if err != nil {
			return all, err
		}
		mu.Lock()
		all = append(all, levelResults...)
		mu.Unlock()
		frontier = next
	}

	return all, nil
}
