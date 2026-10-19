package runner

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"
)

func TestValidURL(t *testing.T) {
	cases := map[string]bool{
		"http://example.com":       true,
		"https://example.com/path": true,
		"ftp://example.com":        false,
		"example.com":              false,
		"":                         false,
		"https://":                 false,
		"  https://example.com  ":  true,
	}
	for in, want := range cases {
		if got := ValidURL(in); got != want {
			t.Errorf("ValidURL(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestResolveURL(t *testing.T) {
	base := "https://example.com/dir/page.html"
	cases := map[string]string{
		"sub/x.html":          "https://example.com/dir/sub/x.html",
		"/abs.html":           "https://example.com/abs.html",
		"https://other.com/y": "https://other.com/y",
		"#frag":               "",
		"javascript:alert(1)": "",
		"mailto:a@b.com":      "",
		"data:text/plain,hi":  "",
	}
	for ref, want := range cases {
		if got := resolveURL(base, ref); got != want {
			t.Errorf("resolveURL(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestTraversalURL(t *testing.T) {
	// FUZZ marker substitution preserves the raw payload.
	if got := traversalURL("https://x.com/?f=FUZZ", "../../etc/passwd"); got != "https://x.com/?f=../../etc/passwd" {
		t.Errorf("FUZZ substitution = %q", got)
	}
	// Non-FUZZ appends to the path directory and drops the query, keeping
	// percent-encoding intact.
	if got := traversalURL("https://x.com/a/b.php", "%2e%2e/etc/passwd"); got != "https://x.com/a/%2e%2e/etc/passwd" {
		t.Errorf("path append = %q", got)
	}
}

func TestBuildTraversalPayloadsInline(t *testing.T) {
	opts := TraversalOptions{
		MinDepth:               1,
		MaxDepth:               2,
		CustomTraversalStrings: []string{"../"},
		CustomPaths:            nil, // no --custom-path: no prefix is prepended
		CustomFiles:            []string{"etc/passwd"},
	}
	// With no prefix and one sequence, depth 1..2: ../etc/passwd, ../../etc/passwd
	got, err := buildTraversalPayloads(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"../etc/passwd",
		"../../etc/passwd",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("payloads = %#v, want %#v", got, want)
	}
}

func TestBuildTraversalPayloadsCustomPathPrefix(t *testing.T) {
	// --custom-path adds prefixed variants in addition to the plain ones.
	opts := TraversalOptions{
		MinDepth:               1,
		MaxDepth:               1,
		CustomTraversalStrings: []string{"../"},
		CustomPaths:            []string{"cgi-bin/"},
		CustomFiles:            []string{"etc/passwd"},
	}
	got, err := buildTraversalPayloads(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"cgi-bin/../etc/passwd", "../etc/passwd"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("payloads = %#v, want %#v", got, want)
	}
}

func TestBuildTraversalPayloadsEncoding(t *testing.T) {
	opts := TraversalOptions{
		MinDepth:               1,
		MaxDepth:               1,
		CustomTraversalStrings: []string{"../"},
		CustomPaths:            []string{"x/"},
		CustomFiles:            []string{"f"},
		Encodings:              []string{"base64"},
	}
	got, err := buildTraversalPayloads(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Each base payload is replaced by its base64 encoding (no plain).
	want := []string{
		base64.StdEncoding.EncodeToString([]byte("x/../f")),
		base64.StdEncoding.EncodeToString([]byte("../f")),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("payloads = %#v, want %#v", got, want)
	}
}

func TestBypassPathVariant(t *testing.T) {
	cases := map[string]string{
		"/":     "https://x.com/admin/",
		"/%2e/": "https://x.com/admin/%2e/",
		"?":     "https://x.com/admin",
	}
	for suffix, want := range cases {
		if got := bypassPathVariant("https://x.com/admin?a=1", suffix); got != want {
			t.Errorf("bypassPathVariant(%q) = %q, want %q", suffix, got, want)
		}
	}
}

func TestBypassPathVariants(t *testing.T) {
	// An up-traversal payload must be tried as suffix, prefix AND wrap so the
	// protected segment is still the effective target.
	got := bypassPathVariants("https://x.com/api/admin", "/..;/")
	want := []string{
		"https://x.com/api/admin/..;/",      // suffix
		"https://x.com/api/..;/admin",       // prefix
		"https://x.com/api/admin/..;/admin", // wrap
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("bypassPathVariants(..;) = %#v, want %#v", got, want)
	}
	// A non-traversal payload gets suffix + prefix only (no wrap).
	got = bypassPathVariants("https://x.com/admin", "/%2f")
	want = []string{
		"https://x.com/admin/%2f",
		"https://x.com/%2fadmin",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("bypassPathVariants(%%2f) = %#v, want %#v", got, want)
	}
	// Query/fragment tricks stay appended to the whole path only.
	got = bypassPathVariants("https://x.com/admin", "?")
	if len(got) != 1 || got[0] != "https://x.com/admin?" {
		t.Errorf("bypassPathVariants(?) = %#v", got)
	}
}

func TestReplaceQueryParam(t *testing.T) {
	got := replaceQueryParam("https://x.com/p?id=1&q=2", "id", "'")
	// url.Values.Encode sorts keys; the id value is percent-encoded.
	want := "https://x.com/p?id=%27&q=2"
	if got != want {
		t.Errorf("replaceQueryParam = %q, want %q", got, want)
	}
}

func TestSafeJoin(t *testing.T) {
	// A traversal attempt in the URL path must not escape the base directory.
	got := safeJoin("base", "evil.com", "/../../etc/passwd")
	want := "base/evil.com/etc/passwd"
	if got != want {
		t.Errorf("safeJoin = %q, want %q", got, want)
	}
}

func TestDedupeEmails(t *testing.T) {
	body := "a@b.com x a@b.com y c@d.net"
	got := dedupeEmails(body)
	want := []string{"a@b.com", "c@d.net"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dedupeEmails = %#v, want %#v", got, want)
	}

	// Real-world noise: trailing dot trimmed, retina-image "addresses" dropped,
	// and case-insensitive de-duplication.
	body = `mailto:john@a.co.uk | sales@a.io. | logo@2x.png | icon@3x.svg | BOB@X.ORG bob@x.org`
	got = dedupeEmails(body)
	want = []string{"john@a.co.uk", "sales@a.io", "BOB@X.ORG"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dedupeEmails(noise) = %#v, want %#v", got, want)
	}
}

func TestApplyEncodings(t *testing.T) {
	// A single encoding replaces the plain payload (no plain is kept).
	// base64 of "'" is "Jw==", of "\"" is "Ig==".
	got := applyEncodings([]string{"'", "\""}, []string{"base64"}, "")
	want := []string{"Jw==", "Ig=="}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	// No encodings: payloads are returned unchanged.
	plain := applyEncodings([]string{"a", "b"}, nil, "")
	if !reflect.DeepEqual(plain, []string{"a", "b"}) {
		t.Errorf("no-encoding should return plain payloads, got %#v", plain)
	}
}

func TestChainEncode(t *testing.T) {
	// base64 alone.
	if got := chainEncode("../x", []string{"base64"}, ""); got != base64.StdEncoding.EncodeToString([]byte("../x")) {
		t.Errorf("base64 chain = %q", got)
	}
	// xor,base64: base64 of the raw xor bytes (pipeline order).
	wantXorB64 := base64.StdEncoding.EncodeToString(xorBytes([]byte("../x"), []byte("k")))
	if got := chainEncode("../x", []string{"xor", "base64"}, "k"); got != wantXorB64 {
		t.Errorf("xor,base64 chain = %q, want %q", got, wantXorB64)
	}
	// A trailing xor is percent-encoded so it is URL-transmittable.
	got := chainEncode("../x", []string{"base64", "xor"}, "k")
	if !strings.HasPrefix(got, "%") {
		t.Errorf("trailing xor should be percent-encoded, got %q", got)
	}
}

func TestBypassHit(t *testing.T) {
	cases := []struct {
		status, base int
		want         bool
	}{
		{200, 403, true},  // success where baseline was forbidden
		{204, 401, true},  // 2xx success
		{404, 403, false}, // just another error, not a bypass
		{500, 403, false}, // server error, not a bypass
		{302, 403, false}, // redirect is not treated as a confirmed bypass
		{200, 200, false}, // baseline was already reachable
	}
	for _, c := range cases {
		if got := bypassHit(c.status, c.base); got != c.want {
			t.Errorf("bypassHit(%d, %d) = %v, want %v", c.status, c.base, got, c.want)
		}
	}
}

func TestBooleanInjectable(t *testing.T) {
	// TRUE resembles baseline (1000), FALSE diverges (300): injectable.
	if !booleanInjectable(1000, 1005, 300, 200, 200, 200) {
		t.Error("expected boolean injectable on length differential")
	}
	// All similar: not injectable (static/non-vulnerable page).
	if booleanInjectable(1000, 1000, 1000, 200, 200, 200) {
		t.Error("did not expect boolean injectable when all responses match")
	}
	// Status differential: TRUE matches baseline, FALSE differs.
	if !booleanInjectable(1000, 1000, 1000, 200, 200, 500) {
		t.Error("expected boolean injectable on status differential")
	}
}

func TestSimilarLen(t *testing.T) {
	if !similarLen(1000, 1005) {
		t.Error("1000 vs 1005 should be similar")
	}
	if similarLen(1000, 300) {
		t.Error("1000 vs 300 should not be similar")
	}
	if !similarLen(10, 20) { // within the 32-byte floor
		t.Error("small lengths within floor should be similar")
	}
}

func TestMatchSignature(t *testing.T) {
	sigs := loadTraversalSignatures("")
	if len(sigs) == 0 {
		t.Skip("default signature wordlist not found from this working dir")
	}
	if got := matchSignature([]byte("root:x:0:0:root:/root:/bin/bash"), sigs); got == "" {
		t.Error("expected /etc/passwd body to match a signature")
	}
	if got := matchSignature([]byte("<html>just a normal page</html>"), sigs); got != "" {
		t.Errorf("did not expect a benign page to match, got %q", got)
	}
}
