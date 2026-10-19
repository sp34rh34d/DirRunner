package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runFuzz(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("fuzz", flag.ExitOnError)
	common := addHTTPFlags(fs, "wordlist/directory.txt", 40, 10*time.Second)
	target := ""
	urlFile := ""
	method := "GET"
	codesRaw := "200,301,302"
	bodyJSON := ""
	bodyForm := ""
	csrfField := ""
	csrfURL := ""
	encode := ""
	xorKey := "DirRunner"
	useBase64 := false
	useURLEncode := false
	fs.StringVar(&target, "url", "", "target URL containing FUZZ or paired with another FUZZ option")
	fs.StringVar(&target, "u", "", "target URL containing FUZZ or paired with another FUZZ option")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	fs.StringVar(&method, "method", "GET", "HTTP method")
	fs.StringVar(&method, "X", "GET", "HTTP method")
	fs.StringVar(&codesRaw, "codes", "200,301,302,401,403", "comma separated status codes to show")
	fs.StringVar(&codesRaw, "c", "200,301,302,401,403", "comma separated status codes to show")
	fs.StringVar(&bodyJSON, "json-body", "", "JSON body containing FUZZ")
	fs.StringVar(&bodyJSON, "b", "", "JSON body containing FUZZ")
	fs.StringVar(&bodyForm, "data", "", "POST form data containing FUZZ, for example user=FUZZ&pass=x")
	fs.StringVar(&bodyForm, "p", "", "POST form data containing FUZZ")
	fs.StringVar(&csrfField, "csrf-field", "", "body field holding an anti-CSRF token; a fresh token+cookie is fetched before each request")
	fs.StringVar(&csrfURL, "csrf-url", "", "page to GET the CSRF token from (default: the target URL)")
	fs.StringVar(&encode, "encode", "", "chain these encodings in order (base64,urlencode,xor); only the encoded result is sent")
	fs.StringVar(&encode, "payload-encoding", "", "alias for --encode")
	fs.StringVar(&xorKey, "xor-key", "DirRunner", "XOR key used with --encode xor")
	fs.BoolVar(&useBase64, "base64", false, "shortcut for --encode base64")
	fs.BoolVar(&useBase64, "B", false, "shortcut for --encode base64")
	fs.BoolVar(&useURLEncode, "urlencode", false, "shortcut: append urlencode to --encode")
	fs.BoolVar(&useURLEncode, "E", false, "shortcut: append urlencode to --encode")
	setModuleUsage(fs, "fuzz", "FUZZ placeholder enumeration for URL, headers, auth, JSON body, or POST form data.", []string{
		"dirrunner fuzz -u 'https://example.com/FUZZ' -w wordlist/passwords.txt --encode base64,urlencode",
		"echo 'https://example.com/FUZZ' | dirrunner fuzz - -w wordlist/passwords.txt",
		"dirrunner fuzz -u https://example.com/login -p 'username=FUZZ&password=fake' -w username.txt",
		"dirrunner fuzz -u https://example.com/login -p 'username=admin&password=FUZZ&csrf=x' --csrf-field csrf -w pass.txt",
		"dirrunner fuzz -u https://example.com/admin --user FUZZ --pass fake -w username.txt",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(true)}, helpGroup{Title: "HTTP flags", Flags: httpHelpFlags()}, helpGroup{Title: "Body flags", Flags: []helpFlag{
		{Names: "--json-body / -b", Use: "JSON", Desc: "JSON body containing FUZZ"},
		{Names: "--data / -p", Use: "FORM", Desc: "POST form data containing FUZZ"},
		{Names: "--csrf-field", Use: "NAME", Desc: "refresh this anti-CSRF token+cookie before each request"},
		{Names: "--csrf-url", Use: "URL", Desc: "page to read the CSRF token from (default: target)"},
	}}, helpGroup{Title: "Encoder flags", Flags: []helpFlag{
		{Names: "--encode", Use: "LIST", Desc: "chain base64,urlencode,xor in order (only encoded sent)"},
		{Names: "--xor-key", Use: "KEY", Desc: "key for --encode xor"},
		{Names: "--base64 / -B", Desc: "shortcut for --encode base64"},
		{Names: "--urlencode / -E", Desc: "shortcut: append urlencode to the chain"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut
	targets, err := targetsFromArgs(target, urlFile, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	encodings := splitCSV(encode)
	if len(encodings) == 0 {
		// Backward-compatible shortcuts (base64 first, then urlencode).
		if useBase64 {
			encodings = append(encodings, "base64")
		}
		if useURLEncode {
			encodings = append(encodings, "urlencode")
		}
	}
	if err := validateEncodings(encodings, "base64", "urlencode", "xor"); err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", targetsDisplay(targets)},
		{"Method", method},
		{"Wordlist", *common.wordlist},
		{"Status codes", codesRaw},
		{"JSON body", activeValue(bodyJSON != "")},
		{"POST data", activeValue(bodyForm != "")},
		{"CSRF field", displayOrDisabled(csrfField)},
		{"Encodings", displayOrDisabled(strings.Join(encodings, ","))},
	}...)
	showOptions("Fuzz Enumeration", common.commonFlags, rows)
	codes, err := parseCodes(codesRaw)
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	httpOpts := common.options()
	httpOpts.Method = method
	results := runTargets(ctx, targets, func(t string) ([]output.Result, error) {
		return runner.RunFuzz(ctx, runner.FuzzOptions{
			TargetURL: t, Wordlist: *common.wordlist, Workers: *common.workers,
			Codes: codes, BodyJSON: bodyJSON, BodyForm: bodyForm,
			CSRFField: csrfField, CSRFURL: csrfURL, HTTP: httpOpts,
			Encodings: encodings, XorKey: xorKey,
		})
	})
	return results, *common.jsonOut, *common.export, nil
}
