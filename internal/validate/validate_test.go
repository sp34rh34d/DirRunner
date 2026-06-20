package validate

import "testing"

func TestDomain(t *testing.T) {
	tests := map[string]bool{
		"example.com":     true,
		"sub.example.com": true,
		"localhost":       true,
		"192.168.1.1":     false,
		"-bad.example":    false,
		"bad-.example":    false,
		"":                false,
	}
	for input, want := range tests {
		if got := Domain(input); got != want {
			t.Fatalf("Domain(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestURL(t *testing.T) {
	tests := map[string]bool{
		"https://example.com":      true,
		"http://example.com/admin": true,
		"ftp://example.com":        false,
		"example.com":              false,
		"https://":                 false,
		"https://example.com/FUZZ": true,
	}
	for input, want := range tests {
		if got := URL(input); got != want {
			t.Fatalf("URL(%q) = %v, want %v", input, got, want)
		}
	}
}
