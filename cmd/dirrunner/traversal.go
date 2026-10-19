package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runTraversal(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("traversal", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 20, 10*time.Second)
	target := ""
	urlFile := ""
	minDepth := 1
	maxDepth := 3
	stringsWordlist := runner.DefaultTraversalStrings
	filesWordlist := runner.DefaultTraversalFiles
	signatureWordlist := runner.DefaultTraversalSignatures
	regx := ""
	customPath := ""
	customString := ""
	customFile := ""
	encode := ""
	xorKey := "DirRunner"
	continueOnSuccess := false
	fs.StringVar(&target, "url", "", "target base URL (supports a FUZZ marker)")
	fs.StringVar(&target, "u", "", "target base URL (supports a FUZZ marker)")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	fs.IntVar(&minDepth, "min-depth", 1, "minimum traversal depth")
	fs.IntVar(&maxDepth, "max-depth", 3, "maximum traversal depth")
	fs.StringVar(&stringsWordlist, "strings-wordlist", runner.DefaultTraversalStrings, "traversal sequence wordlist")
	fs.StringVar(&filesWordlist, "files-wordlist", runner.DefaultTraversalFiles, "file-disclosure wordlist (all payloads, both OSes)")
	fs.StringVar(&signatureWordlist, "signature-wordlist", runner.DefaultTraversalSignatures, "leaked-file content signature wordlist (regex per line)")
	fs.StringVar(&regx, "regx", "", "also report responses matching this regex/word, and show the match")
	fs.StringVar(&customPath, "custom-path", "", "comma separated prefix paths prepended to every payload, e.g. cgi-bin/ (default: no prefix)")
	fs.StringVar(&customString, "custom-traversal-string", "", "comma separated traversal strings (overrides wordlist)")
	fs.StringVar(&customFile, "custom-file", "", "comma separated file-disclosure targets (overrides the files wordlist)")
	fs.StringVar(&encode, "encode", "", "chain these encodings in order (base64,urlencode,xor); only the encoded result is sent")
	fs.StringVar(&encode, "payload-encoding", "", "alias for --encode")
	fs.StringVar(&xorKey, "xor-key", "DirRunner", "XOR key used with --encode xor")
	fs.BoolVar(&continueOnSuccess, "continue-on-success", false, "keep scanning a target after a finding")
	setModuleUsage(fs, "traversal", "Path-traversal / local file disclosure scanner. Payloads are generated from the wordlist/traversal-*.txt lists.", []string{
		"dirrunner traversal -u https://example.com/download.php?file=x",
		"dirrunner traversal -u 'https://example.com/?page=FUZZ'",
		"dirrunner traversal --url-file targets.txt --encode base64,urlencode --max-depth 6",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()}, helpGroup{Title: "Traversal flags", Flags: []helpFlag{
		{Names: "--url / -u", Use: "URL", Desc: "target URL (supports a FUZZ marker)"},
		{Names: "--url-file", Use: "FILE", Desc: "file with target URLs"},
		{Names: "--min-depth / --max-depth", Use: "N", Desc: "traversal depth range"},
		{Names: "--strings-wordlist", Use: "FILE", Desc: "traversal sequence wordlist"},
		{Names: "--files-wordlist", Use: "FILE", Desc: "file-disclosure wordlist (all payloads)"},
		{Names: "--signature-wordlist", Use: "FILE", Desc: "leaked-file content signatures (regex)"},
		{Names: "--regx", Use: "PATTERN", Desc: "also report responses matching this regex/word"},
		{Names: "--custom-path", Use: "LIST", Desc: "prefix paths prepended to payloads (default: none)"},
		{Names: "--custom-traversal-string / --custom-file", Use: "LIST", Desc: "inline overrides (comma separated)"},
		{Names: "--encode", Use: "LIST", Desc: "chain base64,urlencode,xor in order (only encoded sent)"},
		{Names: "--xor-key", Use: "KEY", Desc: "key for --encode xor"},
		{Names: "--continue-on-success", Desc: "keep scanning after a finding"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut

	targets, err := targetsFromArgs(target, urlFile, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	encodings := splitCSV(encode)
	if err := validateEncodings(encodings, "base64", "urlencode", "xor"); err != nil {
		return nil, *common.jsonOut, *common.export, err
	}

	rows := append(httpOptionRows(common), [][2]string{
		{"Targets", fmt.Sprint(len(targets))},
		{"Files wordlist", filesWordlist},
		{"Depth", fmt.Sprintf("%d-%d", minDepth, maxDepth)},
		{"Signatures", signatureWordlist},
		{"Regx", displayOrDisabled(regx)},
		{"Encodings", displayOrDisabled(strings.Join(encodings, ","))},
		{"Continue on success", fmt.Sprint(continueOnSuccess)},
	}...)
	showOptions("Path Traversal", common.commonFlags, rows)

	results, err := runner.RunTraversal(ctx, runner.TraversalOptions{
		Targets:                targets,
		MinDepth:               minDepth,
		MaxDepth:               maxDepth,
		StringsWordlist:        stringsWordlist,
		FilesWordlist:          filesWordlist,
		SignatureWordlist:      signatureWordlist,
		Regx:                   regx,
		CustomPaths:            splitCSV(customPath),
		CustomTraversalStrings: splitCSV(customString),
		CustomFiles:            splitCSV(customFile),
		Encodings:              encodings,
		XorKey:                 xorKey,
		ContinueOnSuccess:      continueOnSuccess,
		Workers:                *common.workers,
		HTTP:                   common.options(),
	})
	return results, *common.jsonOut, *common.export, err
}

// validateEncodings rejects any encoding not in the allowed set.
func validateEncodings(encodings []string, allowed ...string) error {
	allow := map[string]struct{}{}
	for _, a := range allowed {
		allow[a] = struct{}{}
	}
	for _, e := range encodings {
		if _, ok := allow[e]; !ok {
			return fmt.Errorf("invalid encoding %q; allowed: %s", e, strings.Join(allowed, ","))
		}
	}
	return nil
}
