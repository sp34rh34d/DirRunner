package main

import (
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
	cookie            *string
	username          *string
	password          *string
	timeout           *time.Duration
	insecure          *bool
	follow            *bool
	headers           *arrayFlag
	excludeSizes      *string
	excludeSizeRanges *string
	tor               *bool
	torProxy          *string
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
	cookie := ""
	username := ""
	password := ""
	timeout := defaultTimeout
	insecure := false
	follow := false
	excludeSizes := ""
	excludeSizeRanges := ""
	tor := false
	torProxy := "127.0.0.1:9050"
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
	fs.BoolVar(&tor, "tor", false, "route HTTP requests through Tor SOCKS5")
	fs.BoolVar(&tor, "Q", false, "route HTTP requests through Tor SOCKS5")
	fs.StringVar(&torProxy, "tor-proxy", "127.0.0.1:9050", "Tor SOCKS5 proxy address")
	fs.StringVar(&torProxy, "q", "127.0.0.1:9050", "Tor SOCKS5 proxy address")
	var headers arrayFlag
	fs.Var(&headers, "header", "custom header, repeatable: Name: value")
	fs.Var(&headers, "H", "custom header, repeatable: Name: value")
	return httpFlags{
		commonFlags:       commonFlags{wordlist: &wordlist, workers: &workers, jsonOut: &jsonOut, export: &export, verbose: &verbose},
		userAgent:         &userAgent,
		cookie:            &cookie,
		username:          &username,
		password:          &password,
		timeout:           &timeout,
		insecure:          &insecure,
		follow:            &follow,
		headers:           &headers,
		excludeSizes:      &excludeSizes,
		excludeSizeRanges: &excludeSizeRanges,
		tor:               &tor,
		torProxy:          &torProxy,
	}
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
		{"User-Agent", *h.userAgent},
		{"Cookie", activeValue(*h.cookie != "")},
		{"Basic auth", activeValue(*h.username != "" || *h.password != "")},
		{"Timeout", h.timeout.String()},
		{"TLS validation", fmt.Sprint(!*h.insecure)},
		{"Follow redirects", fmt.Sprint(*h.follow)},
		{"Custom headers", fmt.Sprint(len(*h.headers))},
		{"Exclude sizes", displayOrDisabled(*h.excludeSizes)},
		{"Exclude size ranges", displayOrDisabled(*h.excludeSizeRanges)},
		{"Tor", activeValue(*h.tor)},
		{"Tor proxy", torProxyDisplay(*h.tor, *h.torProxy)},
	}
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
		Cookie:          *h.cookie,
		Username:        *h.username,
		Password:        *h.password,
		Timeout:         *h.timeout,
		TLSVerify:       !*h.insecure,
		FollowRedirects: *h.follow,
		Headers:         parseHeaders(*h.headers),
		ExcludeSizes:    parseExcludeSizes(*h.excludeSizes),
		ExcludeRanges:   parseExcludeSizeRanges(*h.excludeSizeRanges),
		Tor:             *h.tor,
		TorProxy:        *h.torProxy,
	}
}

func torProxyDisplay(enabled bool, value string) string {
	if !enabled {
		return "disabled"
	}
	return value
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
