package runner

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExecuteFuzzRequestExcludesResponseSize(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    200,
			ContentLength: 18,
			Header:        make(http.Header),
			Body:          io.NopCloser(strings.NewReader("same-size-response")),
			Request:       req,
		}, nil
	})}

	result, ok := executeFuzzRequest(
		context.Background(),
		client,
		HTTPOptions{Method: http.MethodPost, ExcludeSizes: map[int64]struct{}{18: {}}},
		"https://example.test/login",
		bytes.NewBufferString("username=admin&password=fake"),
		"application/x-www-form-urlencoded",
		map[int]struct{}{200: {}},
	)
	if ok {
		t.Fatalf("expected excluded fuzz result, got %+v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
