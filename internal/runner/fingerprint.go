package runner

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/validate"
)

type FingerprintOptions struct {
	TargetURL string
	HTTP      HTTPOptions
}

var techPatterns = map[string]*regexp.Regexp{
	"Cloudflare": regexp.MustCompile(`(?i)cloudflare`),
	"jQuery":     regexp.MustCompile(`(?i)jquery`),
	"React":      regexp.MustCompile(`(?i)react|__NEXT_DATA__`),
	"Angular":    regexp.MustCompile(`(?i)ng-version|angular`),
	"Vue":        regexp.MustCompile(`(?i)vue(?:\.min)?\.js|data-v-`),
	"WordPress":  regexp.MustCompile(`(?i)wp-content|wp-includes|wordpress`),
	"Drupal":     regexp.MustCompile(`(?i)drupal`),
	"Laravel":    regexp.MustCompile(`(?i)laravel`),
	"PHP":        regexp.MustCompile(`(?i)\bphp\b|x-powered-by:\s*php`),
	"ASP.NET":    regexp.MustCompile(`(?i)asp\.net|x-aspnet`),
	"Express":    regexp.MustCompile(`(?i)express`),
	"Bootstrap":  regexp.MustCompile(`(?i)bootstrap`),
}

func RunFingerprint(ctx context.Context, opts FingerprintOptions) ([]output.Result, error) {
	if !validate.URL(opts.TargetURL) {
		return nil, ErrInvalidTarget("url")
	}
	if opts.HTTP.Timeout == 0 {
		opts.HTTP.Timeout = 15 * time.Second
	}
	opts.HTTP.Method = http.MethodGet
	client := NewHTTPClient(opts.HTTP)
	req, err := NewRequest(ctx, http.MethodGet, opts.TargetURL, nil, opts.HTTP)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	headers := make([]string, 0, len(resp.Header))
	for k, values := range resp.Header {
		headers = append(headers, k+": "+strings.Join(values, ", "))
	}
	sort.Strings(headers)
	haystack := strings.Join(headers, "\n") + "\n" + string(bodyBytes)

	results := []output.Result{{
		Type:   "fingerprint",
		Target: opts.TargetURL,
		URL:    opts.TargetURL,
		Status: resp.StatusCode,
		Size:   int64(len(bodyBytes)),
		Value:  "response",
	}}
	for name, pattern := range techPatterns {
		if pattern.MatchString(haystack) {
			results = append(results, output.Result{
				Type:   "technology",
				Target: opts.TargetURL,
				URL:    opts.TargetURL,
				Status: resp.StatusCode,
				Size:   int64(len(bodyBytes)),
				Value:  name,
			})
		}
	}
	for _, header := range headers {
		results = append(results, output.Result{
			Type:   "header",
			Target: opts.TargetURL,
			URL:    opts.TargetURL,
			Status: resp.StatusCode,
			Size:   int64(len(bodyBytes)),
			Value:  header,
		})
	}
	return results, nil
}
