package runner

import "testing"

func TestJoinURL(t *testing.T) {
	got := JoinURL("https://example.com/base/", "/admin", true)
	want := "https://example.com/base/admin/"
	if got != want {
		t.Fatalf("JoinURL() = %q, want %q", got, want)
	}
}

func TestStatusSet(t *testing.T) {
	set := StatusSet([]int{200, 301, 999, 0})
	if _, ok := set[200]; !ok {
		t.Fatal("expected 200 to be present")
	}
	if _, ok := set[999]; ok {
		t.Fatal("did not expect invalid status to be present")
	}
}
