package runner

import (
	"context"
	"net/http"
	"strings"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/validate"
)

type VHostOptions struct {
	TargetURL string
	Domain    string
	Wordlist  string
	Workers   int
	HTTP      HTTPOptions
}

func RunVHost(ctx context.Context, opts VHostOptions) ([]output.Result, error) {
	if !validate.URL(opts.TargetURL) {
		return nil, ErrInvalidTarget("url")
	}
	if !validate.Domain(opts.Domain) {
		return nil, ErrInvalidTarget("domain")
	}
	if opts.HTTP.Timeout == 0 {
		opts.HTTP.Timeout = 2 * time.Second
	}
	opts.HTTP.Method = http.MethodGet
	client := NewHTTPClient(opts.HTTP)
	codes := StatusSet([]int{http.StatusOK})

	return RunWorkers(ctx, opts.Wordlist, opts.Workers, func(ctx context.Context, word string) (output.Result, bool) {
		host := strings.Trim(strings.TrimSpace(word)+"."+opts.Domain, ".")
		httpOpts := opts.HTTP
		httpOpts.Headers = cloneHeaders(opts.HTTP.Headers)
		httpOpts.Headers["Host"] = host
		result, ok := executeHTTPRequest(ctx, client, httpOpts, opts.TargetURL, codes)
		if !ok {
			return output.Result{}, false
		}
		result.Type = "vhost"
		result.Target = opts.TargetURL
		result.Host = host
		return result, true
	})
}

func cloneHeaders(headers map[string]string) map[string]string {
	copied := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		copied[k] = v
	}
	return copied
}
