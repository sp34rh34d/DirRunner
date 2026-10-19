package runner

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"

	"dirrunner/internal/output"
)

// Default wordlist paths for the traversal module. The sequences and the
// per-OS file-disclosure lists live under wordlist/ so they can be tuned
// without recompiling.
const (
	DefaultTraversalStrings    = "wordlist/traversal-strings.txt"
	DefaultTraversalFiles      = "wordlist/traversal-files.txt"
	DefaultTraversalSignatures = "wordlist/traversal-signatures.txt"
)

type TraversalOptions struct {
	Targets                []string
	MinDepth               int
	MaxDepth               int
	StringsWordlist        string
	FilesWordlist          string
	SignatureWordlist      string
	Regx                   string
	CustomPaths            []string
	CustomTraversalStrings []string
	CustomFiles            []string
	Encodings              []string
	XorKey                 string
	ContinueOnSuccess      bool
	Workers                int
	HTTP                   HTTPOptions
}

// RunTraversal probes each target with generated path-traversal payloads and
// reports responses that differ from the target's baseline body length,
// mirroring WebRunner's traversal module.
func RunTraversal(ctx context.Context, opts TraversalOptions) ([]output.Result, error) {
	if opts.MinDepth < 1 {
		opts.MinDepth = 1
	}
	if opts.MaxDepth < opts.MinDepth {
		return nil, fmt.Errorf("--min-depth cannot be greater than --max-depth")
	}
	payloads, err := buildTraversalPayloads(opts)
	if err != nil {
		return nil, err
	}
	if len(payloads) == 0 {
		return nil, fmt.Errorf("no traversal payloads were generated; check the traversal wordlists")
	}
	output.Info("%d traversal payloads generated", len(payloads))

	// Detection is signature-based: a hit is confirmed when the response body
	// contains a leaked-file fingerprint (root:...:0:0, [boot loader], a private
	// key, ...). --regx adds an optional ad-hoc content check on top.
	signatures := loadTraversalSignatures(opts.SignatureWordlist)
	if len(signatures) == 0 {
		output.Warn("no content signatures loaded (check wordlist/traversal-signatures.txt)")
	} else {
		output.Info("%d content signatures loaded", len(signatures))
	}
	var re *regexp.Regexp
	if opts.Regx != "" {
		compiled, err := regexp.Compile(opts.Regx)
		if err != nil {
			return nil, fmt.Errorf("invalid --regx pattern: %w", err)
		}
		re = compiled
		output.Info("regx filter active: %q", opts.Regx)
	}
	if len(signatures) == 0 && re == nil {
		return nil, fmt.Errorf("no detector available: provide content signatures or pass --regx")
	}

	client := NewHTTPClient(opts.HTTP)
	var all []output.Result
	for _, target := range opts.Targets {
		if ctx.Err() != nil {
			break
		}
		output.Info("scanning %s", target)

		// A cancelable context lets a finding stop the whole pool at once when
		// --continue-on-success is off, instead of only skipping not-yet-started
		// jobs (which, with many workers, still lets dozens of in-flight requests
		// complete and report).
		poolCtx, cancel := context.WithCancel(ctx)
		var found atomic.Bool
		results, err := RunPool(poolCtx, "traversal", payloads, opts.Workers, func(ctx context.Context, payload string) []output.Result {
			if !opts.ContinueOnSuccess && found.Load() {
				return nil
			}
			testURL := traversalURL(target, payload)
			if testURL == "" {
				return nil
			}
			output.Debug("testing %s", testURL)
			status, body, size, err := fetchDetect(ctx, client, opts.HTTP, "GET", testURL, nil)
			if err != nil {
				output.Debug("error %s: %v", testURL, err)
				return nil
			}
			// Signature match (confirms a real file leak on any status), then the
			// optional --regx content match.
			reason := ""
			if sig := matchSignature(body, signatures); sig != "" {
				reason = "signature: " + sig
			} else if re != nil {
				if m := re.Find(body); m != nil {
					reason = "regx: " + truncateMatch(string(m))
				}
			}
			if reason == "" {
				return nil
			}
			found.Store(true)
			if !opts.ContinueOnSuccess {
				cancel()
			}
			return []output.Result{{
				Type:   "traversal",
				Target: target,
				Method: "GET",
				URL:    testURL,
				Status: status,
				Size:   size,
				Value:  payload + " | " + reason,
			}}
		})
		cancel()
		if err != nil {
			return all, err
		}
		all = append(all, results...)
	}
	return all, nil
}

// truncateMatch shortens a regx match for display.
func truncateMatch(s string) string {
	const max = 80
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

// loadTraversalSignatures compiles the signature wordlist as case-insensitive
// regexes. A missing file is not fatal: it disables signature matching.
func loadTraversalSignatures(path string) []*regexp.Regexp {
	lines, err := LoadWordlist(firstNonEmpty(path, DefaultTraversalSignatures))
	if err != nil {
		return nil
	}
	patterns := make([]*regexp.Regexp, 0, len(lines))
	for _, line := range lines {
		if re, err := regexp.Compile("(?i)" + line); err == nil {
			patterns = append(patterns, re)
		} else {
			output.Warn("skipping invalid traversal signature %q: %v", line, err)
		}
	}
	return patterns
}

// matchSignature returns the first signature that matches body, or "".
func matchSignature(body []byte, signatures []*regexp.Regexp) string {
	for _, re := range signatures {
		if loc := re.Find(body); loc != nil {
			return re.String()
		}
	}
	return ""
}

// traversalURL builds the request URL for a payload. When the target contains a
// FUZZ marker the payload is substituted there; otherwise it is appended to the
// target's directory. The URL is assembled by string concatenation so that
// percent-encoded payload bytes are preserved rather than re-encoded.
func traversalURL(target, payload string) string {
	if strings.Contains(target, "FUZZ") {
		return strings.ReplaceAll(target, "FUZZ", payload)
	}
	u, err := url.Parse(target)
	if err != nil {
		return ""
	}
	dir := u.Path
	if idx := strings.LastIndex(dir, "/"); idx >= 0 {
		dir = dir[:idx+1]
	} else {
		dir = "/"
	}
	return fmt.Sprintf("%s://%s%s%s", u.Scheme, u.Host, dir, payload)
}

func buildTraversalPayloads(opts TraversalOptions) ([]string, error) {
	strs := opts.CustomTraversalStrings
	if len(strs) == 0 {
		loaded, err := LoadWordlist(firstNonEmpty(opts.StringsWordlist, DefaultTraversalStrings))
		if err != nil {
			return nil, fmt.Errorf("loading traversal strings: %w", err)
		}
		strs = loaded
	}

	// Prefix paths are only applied when --custom-path is set. By default no
	// prefix is prepended, so only the plain "<sequence><file>" payloads are
	// generated (prefixes like cgi-bin/ are an opt-in, not a default cost).
	paths := opts.CustomPaths

	files := opts.CustomFiles
	if len(files) == 0 {
		loaded, err := LoadWordlist(firstNonEmpty(opts.FilesWordlist, DefaultTraversalFiles))
		if err != nil {
			return nil, fmt.Errorf("loading traversal files: %w", err)
		}
		files = loaded
	}

	seen := map[string]struct{}{}
	var payloads []string
	add := func(p string) {
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		payloads = append(payloads, p)
	}

	for depth := opts.MinDepth; depth <= opts.MaxDepth; depth++ {
		for _, o := range paths {
			for _, s := range strs {
				for _, f := range files {
					add(o + strings.Repeat(s, depth) + f)
				}
			}
		}
		for _, s := range strs {
			for _, f := range files {
				add(strings.Repeat(s, depth) + f)
			}
		}
	}

	// Append encoded variants (base64, urlencode, xor) of the base payloads.
	return applyEncodings(payloads, opts.Encodings, opts.XorKey), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// percentEncodeAll percent-encodes every byte, matching Python's
// quote_from_bytes(..., safe="").
func percentEncodeAll(b []byte) string {
	var sb strings.Builder
	sb.Grow(len(b) * 3)
	const hex = "0123456789ABCDEF"
	for _, c := range b {
		sb.WriteByte('%')
		sb.WriteByte(hex[c>>4])
		sb.WriteByte(hex[c&0x0f])
	}
	return sb.String()
}
