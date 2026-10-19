package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

type arrayFlag []string

func (a *arrayFlag) String() string { return strings.Join(*a, ",") }
func (a *arrayFlag) Set(value string) error {
	*a = append(*a, value)
	return nil
}

type commonFlags struct {
	wordlist *string
	workers  *int
	jsonOut  *bool
	export   *string
	verbose  *bool
}

type httpFlags struct {
	commonFlags
	userAgent         *string
	randomAgent       *bool
	cookie            *string
	username          *string
	password          *string
	timeout           *time.Duration
	insecure          *bool
	follow            *bool
	headers           *arrayFlag
	excludeSizes      *string
	excludeSizeRanges *string
	proxy             *string
}

func addCommonFlags(fs *flag.FlagSet, defaultWordlist string, defaultWorkers int) commonFlags {
	wordlist := defaultWordlist
	workers := defaultWorkers
	jsonOut := false
	export := ""
	verbose := false
	fs.StringVar(&wordlist, "wordlist", defaultWordlist, "wordlist path")
	fs.StringVar(&wordlist, "w", defaultWordlist, "wordlist path")
	fs.IntVar(&workers, "threads", defaultWorkers, "concurrent workers")
	fs.IntVar(&workers, "t", defaultWorkers, "concurrent workers")
	fs.BoolVar(&jsonOut, "json", false, "print JSON output")
	fs.BoolVar(&jsonOut, "j", false, "print JSON output")
	fs.StringVar(&export, "export", "", "write results to a file")
	fs.StringVar(&export, "o", "", "write results to a file")
	fs.BoolVar(&verbose, "verbose", false, "print verbose progress")
	fs.BoolVar(&verbose, "v", false, "print verbose progress")
	return commonFlags{wordlist: &wordlist, workers: &workers, jsonOut: &jsonOut, export: &export, verbose: &verbose}
}

func addHTTPFlags(fs *flag.FlagSet, defaultWordlist string, defaultWorkers int, defaultTimeout time.Duration) httpFlags {
	wordlist := defaultWordlist
	workers := defaultWorkers
	jsonOut := false
	export := ""
	verbose := false
	userAgent := runner.UserAgent
	randomAgent := false
	cookie := ""
	username := ""
	password := ""
	timeout := defaultTimeout
	insecure := false
	follow := false
	excludeSizes := ""
	excludeSizeRanges := ""
	proxy := ""
	fs.StringVar(&wordlist, "wordlist", defaultWordlist, "wordlist path")
	fs.StringVar(&wordlist, "w", defaultWordlist, "wordlist path")
	fs.IntVar(&workers, "threads", defaultWorkers, "concurrent workers")
	fs.IntVar(&workers, "t", defaultWorkers, "concurrent workers")
	fs.BoolVar(&jsonOut, "json", false, "print JSON output")
	fs.BoolVar(&jsonOut, "j", false, "print JSON output")
	fs.StringVar(&export, "export", "", "write results to a file")
	fs.StringVar(&export, "o", "", "write results to a file")
	fs.BoolVar(&verbose, "verbose", false, "print verbose progress")
	fs.BoolVar(&verbose, "v", false, "print verbose progress")
	fs.StringVar(&userAgent, "user-agent", runner.UserAgent, "User-Agent header")
	fs.StringVar(&userAgent, "A", runner.UserAgent, "User-Agent header")
	fs.BoolVar(&randomAgent, "random-agent", false, "use a random User-Agent per request")
	fs.BoolVar(&randomAgent, "a", false, "use a random User-Agent per request")
	fs.StringVar(&cookie, "cookie", "", "Cookie header")
	fs.StringVar(&cookie, "C", "", "Cookie header")
	fs.StringVar(&username, "username", "", "basic auth username")
	fs.StringVar(&username, "user", "", "basic auth username")
	fs.StringVar(&username, "U", "", "basic auth username")
	fs.StringVar(&password, "password", "", "basic auth password")
	fs.StringVar(&password, "pass", "", "basic auth password")
	fs.StringVar(&password, "P", "", "basic auth password")
	fs.DurationVar(&timeout, "timeout", defaultTimeout, "request timeout")
	fs.DurationVar(&timeout, "T", defaultTimeout, "request timeout")
	fs.BoolVar(&insecure, "insecure", false, "disable TLS certificate validation")
	fs.BoolVar(&insecure, "k", false, "disable TLS certificate validation")
	fs.BoolVar(&follow, "follow-redirects", false, "follow redirects")
	fs.BoolVar(&follow, "L", false, "follow redirects")
	fs.StringVar(&excludeSizes, "exclude-size", "", "comma separated response sizes to exclude")
	fs.StringVar(&excludeSizes, "S", "", "comma separated response sizes to exclude")
	fs.StringVar(&excludeSizeRanges, "exclude-size-range", "", "comma separated inclusive ranges to exclude, for example 100-200,1000-1200")
	fs.StringVar(&excludeSizeRanges, "R", "", "comma separated inclusive ranges to exclude")
	fs.StringVar(&proxy, "proxy", "", "proxy URL: http://, https://, socks5:// or socks5h:// (Tor: socks5h://127.0.0.1:9050)")
	fs.StringVar(&proxy, "x", "", "proxy URL (alias for --proxy)")
	var headers arrayFlag
	fs.Var(&headers, "header", "custom header, repeatable: Name: value")
	fs.Var(&headers, "H", "custom header, repeatable: Name: value")
	return httpFlags{
		commonFlags:       commonFlags{wordlist: &wordlist, workers: &workers, jsonOut: &jsonOut, export: &export, verbose: &verbose},
		userAgent:         &userAgent,
		randomAgent:       &randomAgent,
		cookie:            &cookie,
		username:          &username,
		password:          &password,
		timeout:           &timeout,
		insecure:          &insecure,
		follow:            &follow,
		headers:           &headers,
		excludeSizes:      &excludeSizes,
		excludeSizeRanges: &excludeSizeRanges,
		proxy:             &proxy,
	}
}

// runTargets runs the given per-target function over every target, streaming
// each one's results live and aggregating them. A failure on one target is
// reported and skipped so the rest of the list still runs.
func runTargets(ctx context.Context, targets []string, run func(string) ([]output.Result, error)) []output.Result {
	var all []output.Result
	for _, target := range targets {
		if ctx.Err() != nil {
			break
		}
		results, err := run(target)
		if err != nil {
			output.Warn("failed for %s: %v", target, err)
			continue
		}
		all = append(all, results...)
	}
	return all
}

// targetsDisplay renders the target count for the options menu.
func targetsDisplay(targets []string) string {
	if len(targets) == 1 {
		return targets[0]
	}
	return fmt.Sprintf("%d URLs", len(targets))
}

func showOptions(title string, common commonFlags, rows [][2]string) {
	output.Verbose = *common.verbose
	base := [][2]string{
		{"Threads", fmt.Sprint(*common.workers)},
		{"JSON output", fmt.Sprint(*common.jsonOut)},
		{"Export", *common.export},
		{"Verbose", fmt.Sprint(*common.verbose)},
	}
	rows = append(base, rows...)
	if err := output.OptionsMenu(os.Stderr, title, rows); err != nil {
		output.Warn("could not print options menu: %v", err)
	}
}

func httpOptionRows(h httpFlags) [][2]string {
	return [][2]string{
		{"User-Agent", userAgentDisplay(*h.randomAgent, *h.userAgent)},
		{"Cookie", activeValue(*h.cookie != "")},
		{"Basic auth", activeValue(*h.username != "" || *h.password != "")},
		{"Timeout", h.timeout.String()},
		{"TLS validation", fmt.Sprint(!*h.insecure)},
		{"Follow redirects", fmt.Sprint(*h.follow)},
		{"Custom headers", fmt.Sprint(len(*h.headers))},
		{"Exclude sizes", displayOrDisabled(*h.excludeSizes)},
		{"Exclude size ranges", displayOrDisabled(*h.excludeSizeRanges)},
		{"Proxy", displayOrDisabled(*h.proxy)},
	}
}

func userAgentDisplay(random bool, value string) string {
	if random {
		return "random (rotating per request)"
	}
	return value
}

func activeValue(active bool) string {
	if active {
		return "enabled"
	}
	return "disabled"
}

func (h httpFlags) options() runner.HTTPOptions {
	return runner.HTTPOptions{
		UserAgent:       *h.userAgent,
		RandomUserAgent: *h.randomAgent,
		Cookie:          *h.cookie,
		Username:        *h.username,
		Password:        *h.password,
		Timeout:         *h.timeout,
		TLSVerify:       !*h.insecure,
		FollowRedirects: *h.follow,
		Headers:         parseHeaders(*h.headers),
		ExcludeSizes:    parseExcludeSizes(*h.excludeSizes),
		ExcludeRanges:   parseExcludeSizeRanges(*h.excludeSizeRanges),
		Proxy:           *h.proxy,
	}
}

func displayOrDisabled(value string) string {
	if strings.TrimSpace(value) == "" {
		return "disabled"
	}
	return value
}

func parseCodes(value string) ([]int, error) {
	var codes []int
	for _, part := range splitCSV(value) {
		code, err := strconv.Atoi(part)
		if err != nil || code < 100 || code > 599 {
			return nil, fmt.Errorf("invalid status code: %s", part)
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if clean := strings.TrimSpace(part); clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func parseHeaders(values []string) map[string]string {
	headers := map[string]string{}
	for _, value := range values {
		name, val, ok := strings.Cut(value, ":")
		if ok && strings.TrimSpace(name) != "" {
			headers[strings.TrimSpace(name)] = strings.TrimSpace(val)
		}
	}
	return headers
}

func parseExcludeSizes(value string) map[int64]struct{} {
	sizes := map[int64]struct{}{}
	for _, part := range splitCSV(value) {
		size, err := strconv.ParseInt(part, 10, 64)
		if err == nil && size >= 0 {
			sizes[size] = struct{}{}
		}
	}
	return sizes
}

func parseExcludeSizeRanges(value string) []runner.SizeRange {
	var ranges []runner.SizeRange
	for _, part := range splitCSV(value) {
		minRaw, maxRaw, ok := strings.Cut(part, "-")
		if !ok {
			continue
		}
		min, minErr := strconv.ParseInt(strings.TrimSpace(minRaw), 10, 64)
		max, maxErr := strconv.ParseInt(strings.TrimSpace(maxRaw), 10, 64)
		if minErr != nil || maxErr != nil || min < 0 || max < 0 {
			continue
		}
		if min > max {
			min, max = max, min
		}
		ranges = append(ranges, runner.SizeRange{Min: min, Max: max})
	}
	return ranges
}
