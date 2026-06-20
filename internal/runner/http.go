package runner

import (
	"context"
	"net/http"
	"strings"

	"dirrunner/internal/output"
)

func executeHTTPRequest(ctx context.Context, client *http.Client, opts HTTPOptions, url string, codes map[int]struct{}) (output.Result, bool) {
	req, err := NewRequest(ctx, opts.Method, url, nil, opts)
	if err != nil {
		return output.Result{}, false
	}
	resp, err := client.Do(req)
	if err != nil {
		return output.Result{}, false
	}
	defer resp.Body.Close()
	if _, wanted := codes[resp.StatusCode]; !wanted {
		return output.Result{}, false
	}
	size := ResponseSize(resp)
	if ExcludedSize(size, opts) {
		output.Debug("excluded %s %s status=%d size=%d", strings.ToUpper(req.Method), url, resp.StatusCode, size)
		return output.Result{}, false
	}
	output.Debug("matched %s %s status=%d size=%d", strings.ToUpper(req.Method), url, resp.StatusCode, size)
	return output.Result{
		Method:   strings.ToUpper(req.Method),
		Status:   resp.StatusCode,
		Size:     size,
		Location: resp.Header.Get("Location"),
	}, true
}

func wildcardStatus(ctx context.Context, client *http.Client, opts HTTPOptions, target string, codes map[int]struct{}) (bool, error) {
	url := JoinURL(target, RandomToken(), true)
	result, ok := executeHTTPRequest(ctx, client, opts, url, codes)
	return ok && result.Status > 0, nil
}
