package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"dirrunner/internal/output"
)

// TestContinuousScanRecursesInParallel verifies that recursive scanning
// discovers nested directories (depth > 1) via the parallel dispatcher.
func TestContinuousScanRecursesInParallel(t *testing.T) {
	real := map[string]bool{
		"/admin/":        true,
		"/admin/users/":  true,
		"/admin/config/": true,
		"/blog/":         true,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if real[r.URL.Path] {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	wl := filepath.Join(dir, "words.txt")
	if err := os.WriteFile(wl, []byte("admin\nblog\nusers\nconfig\nnope1\nnope2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output.StreamResults = false
	results, err := RunScan(context.Background(), ScanOptions{
		Target:            srv.URL,
		DirectoryWordlist: wl,
		Workers:           8,
		Codes:             []int{http.StatusOK},
		Recursive:         true,
		MaxDepth:          2,
		IncludeDirs:       true,
		SkipWildcardCheck: true,
	})
	if err != nil {
		t.Fatalf("RunScan() error = %v", err)
	}

	got := make([]string, 0, len(results))
	for _, r := range results {
		got = append(got, r.Path)
	}
	sort.Strings(got)
	want := []string{"/admin/", "/admin/config/", "/admin/users/", "/blog/"}
	if len(got) != len(want) {
		t.Fatalf("found %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("found %v, want %v", got, want)
		}
	}
}
