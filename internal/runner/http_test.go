package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequestSetsHostFromHeader(t *testing.T) {
	req, err := NewRequest(context.Background(), http.MethodGet, "https://example.test/", nil, HTTPOptions{
		Headers: map[string]string{"Host": "vhost.example.com"},
	})
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if req.Host != "vhost.example.com" {
		t.Fatalf("req.Host = %q, want %q", req.Host, "vhost.example.com")
	}
	if got := req.Header.Get("Host"); got != "" {
		t.Fatalf("Host should not be stored in Header, got %q", got)
	}
}

func TestVHostHostHeaderReachesServer(t *testing.T) {
	var gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	opts := HTTPOptions{Headers: map[string]string{"Host": "secret.internal"}}
	client := NewHTTPClient(opts)
	codes := StatusSet([]int{http.StatusOK})
	if _, ok := executeHTTPRequest(context.Background(), client, opts, srv.URL, codes); !ok {
		t.Fatal("expected request to match 200")
	}
	if gotHost != "secret.internal" {
		t.Fatalf("server saw Host %q, want %q", gotHost, "secret.internal")
	}
}
