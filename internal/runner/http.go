package runner

import (
	"context"
	"net/http"
	"strings"

	"dirrunner/internal/output"
)

func executeHTTPRequest(ctx context.Context, client *http.Client, opts HTTPOptions, url string, codes map[int]struct{}) (output.Result, bool) {
	result, _, matched, _ := doHTTPRequest(ctx, client, opts, url, codes)
	return result, matched
}

// doHTTPRequest performs one request and reports, in addition to a matched
// result, the observed status code and any transport error. Scans use these to
// build a summary so a run that finds nothing can still explain why (all
// requests errored, everything was 404, a WAF returned 403, ...).
func doHTTPRequest(ctx context.Context, client *http.Client, opts HTTPOptions, url string, codes map[int]struct{}) (output.Result, int, bool, error) {
	req, err := NewRequest(ctx, opts.Method, url, nil, opts)
	if err != nil {
		return output.Result{}, 0, false, err
	}
	method := strings.ToUpper(req.Method)
	resp, err := client.Do(req)
	if err != nil {
		output.Debug("error %s %s: %v", method, url, err)
		return output.Result{}, 0, false, err
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	if _, wanted := codes[status]; !wanted {
		output.Debug("skip %s %s status=%d", method, url, status)
		return output.Result{}, status, false, nil
	}
	size := ResponseSize(resp)
	if ExcludedSize(size, opts) {
		output.Debug("excluded %s %s status=%d size=%d", method, url, status, size)
		return output.Result{}, status, false, nil
	}
	output.Debug("matched %s %s status=%d size=%d", method, url, status, size)
	return output.Result{
		Method:   method,
		Status:   status,
		Size:     size,
		Location: resp.Header.Get("Location"),
	}, status, true, nil
}

func wildcardStatus(ctx context.Context, client *http.Client, opts HTTPOptions, target string, codes map[int]struct{}) (bool, error) {
	url := JoinURL(target, RandomToken(), true)
	result, ok := executeHTTPRequest(ctx, client, opts, url, codes)
	return ok && result.Status > 0, nil
}
