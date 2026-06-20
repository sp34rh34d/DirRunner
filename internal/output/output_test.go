package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintResultsShowsFuzzValueWhenURLIsUnchanged(t *testing.T) {
	var buf bytes.Buffer
	err := PrintResults(&buf, []Result{{
		Type:   "fuzz",
		Target: "https://example.test/login",
		Method: "POST",
		URL:    "https://example.test/login",
		Status: 200,
		Size:   1234,
		Value:  "admin",
	}}, false)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "admin") {
		t.Fatalf("expected output to include fuzz value, got %q", out)
	}
	if strings.Contains(out, "https://example.test/login") {
		t.Fatalf("expected output to omit unchanged fuzz URL, got %q", out)
	}
}
