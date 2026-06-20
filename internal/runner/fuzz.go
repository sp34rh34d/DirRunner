package runner

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/validate"
)

type FuzzOptions struct {
	TargetURL string
	Wordlist  string
	Workers   int
	Codes     []int
	BodyJSON  string
	BodyForm  string
	HTTP      HTTPOptions
	Encoders  []FuzzEncoder
}

type FuzzEncoder string

const (
	FuzzEncoderBase64    FuzzEncoder = "base64"
	FuzzEncoderURLEncode FuzzEncoder = "urlencode"
)

func RunFuzz(ctx context.Context, opts FuzzOptions) ([]output.Result, error) {
	if !validate.URL(strings.ReplaceAll(opts.TargetURL, "FUZZ", "test")) {
		return nil, ErrInvalidTarget("url")
	}
	if opts.HTTP.Timeout == 0 {
		opts.HTTP.Timeout = 10 * time.Second
	}
	if opts.HTTP.Method == "" {
		opts.HTTP.Method = http.MethodGet
	}
	if opts.BodyJSON != "" && opts.BodyForm != "" {
		return nil, ErrInvalidTarget("body: use json-body or data, not both")
	}
	if opts.BodyJSON != "" {
		opts.HTTP.Method = http.MethodPost
		if !json.Valid([]byte(strings.ReplaceAll(opts.BodyJSON, "FUZZ", "test"))) {
			return nil, ErrInvalidTarget("json body")
		}
	}
	if opts.BodyForm != "" {
		opts.HTTP.Method = http.MethodPost
	}
	if !opts.containsFuzz() {
		return nil, ErrInvalidTarget("FUZZ marker")
	}

	client := NewHTTPClient(opts.HTTP)
	codes := StatusSet(opts.Codes)
	return RunWorkers(ctx, opts.Wordlist, opts.Workers, func(ctx context.Context, word string) (output.Result, bool) {
		encodedWord := opts.encodeFuzzValue(word)
		httpOpts := opts.HTTP
		httpOpts.UserAgent = strings.ReplaceAll(httpOpts.UserAgent, "FUZZ", encodedWord)
		httpOpts.Cookie = strings.ReplaceAll(httpOpts.Cookie, "FUZZ", encodedWord)
		httpOpts.Username = strings.ReplaceAll(httpOpts.Username, "FUZZ", encodedWord)
		httpOpts.Password = strings.ReplaceAll(httpOpts.Password, "FUZZ", encodedWord)
		httpOpts.Headers = make(map[string]string, len(opts.HTTP.Headers))
		for k, v := range opts.HTTP.Headers {
			httpOpts.Headers[strings.ReplaceAll(k, "FUZZ", encodedWord)] = strings.ReplaceAll(v, "FUZZ", encodedWord)
		}

		rawURL := strings.ReplaceAll(opts.TargetURL, "FUZZ", encodedWord)
		var body io.Reader
		contentType := ""
		if opts.BodyJSON != "" {
			body = bytes.NewBufferString(strings.ReplaceAll(opts.BodyJSON, "FUZZ", encodedWord))
			contentType = "application/json"
		}
		if opts.BodyForm != "" {
			body = bytes.NewBufferString(strings.ReplaceAll(opts.BodyForm, "FUZZ", encodedWord))
			contentType = "application/x-www-form-urlencoded"
		}
		result, ok := executeFuzzRequest(ctx, client, httpOpts, rawURL, body, contentType, codes)
		if !ok {
			return output.Result{}, false
		}
		result.Type = "fuzz"
		result.Target = opts.TargetURL
		result.URL = rawURL
		result.Value = encodedWord
		return result, true
	})
}

func (opts FuzzOptions) encodeFuzzValue(value string) string {
	encoded := value
	for _, encoder := range opts.Encoders {
		switch encoder {
		case FuzzEncoderBase64:
			encoded = base64.StdEncoding.EncodeToString([]byte(encoded))
		case FuzzEncoderURLEncode:
			encoded = url.QueryEscape(encoded)
		}
	}
	return encoded
}

func (opts FuzzOptions) containsFuzz() bool {
	if strings.Contains(opts.TargetURL, "FUZZ") ||
		strings.Contains(opts.BodyJSON, "FUZZ") ||
		strings.Contains(opts.BodyForm, "FUZZ") ||
		strings.Contains(opts.HTTP.UserAgent, "FUZZ") ||
		strings.Contains(opts.HTTP.Cookie, "FUZZ") ||
		strings.Contains(opts.HTTP.Username, "FUZZ") ||
		strings.Contains(opts.HTTP.Password, "FUZZ") {
		return true
	}
	for k, v := range opts.HTTP.Headers {
		if strings.Contains(k, "FUZZ") || strings.Contains(v, "FUZZ") {
			return true
		}
	}
	return false
}

func executeFuzzRequest(ctx context.Context, client *http.Client, opts HTTPOptions, rawURL string, body io.Reader, contentType string, codes map[int]struct{}) (output.Result, bool) {
	req, err := NewRequest(ctx, opts.Method, rawURL, body, opts)
	if err != nil {
		return output.Result{}, false
	}
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
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
		output.Debug("excluded %s %s status=%d size=%d", strings.ToUpper(req.Method), rawURL, resp.StatusCode, size)
		return output.Result{}, false
	}
	output.Debug("matched %s %s status=%d size=%d", strings.ToUpper(req.Method), rawURL, resp.StatusCode, size)
	return output.Result{
		Method:   strings.ToUpper(req.Method),
		Status:   resp.StatusCode,
		Size:     size,
		Location: resp.Header.Get("Location"),
	}, true
}
