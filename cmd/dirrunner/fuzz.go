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
	common := addHTTPFlags(fs, "wordlist/directory.txt", 10, 10*time.Second)
	target := ""
	method := "GET"
	codesRaw := "200,301,302"
	bodyJSON := ""
	bodyForm := ""
	useBase64 := false
	useURLEncode := false
	fs.StringVar(&target, "url", "", "target URL containing FUZZ or paired with another FUZZ option")
	fs.StringVar(&target, "u", "", "target URL containing FUZZ or paired with another FUZZ option")
	fs.StringVar(&method, "method", "GET", "HTTP method")
	fs.StringVar(&method, "X", "GET", "HTTP method")
	fs.StringVar(&codesRaw, "codes", "200,301,302", "comma separated status codes to show")
	fs.StringVar(&codesRaw, "c", "200,301,302", "comma separated status codes to show")
	fs.StringVar(&bodyJSON, "json-body", "", "JSON body containing FUZZ")
	fs.StringVar(&bodyJSON, "b", "", "JSON body containing FUZZ")
	fs.StringVar(&bodyForm, "data", "", "POST form data containing FUZZ, for example user=FUZZ&pass=x")
	fs.StringVar(&bodyForm, "p", "", "POST form data containing FUZZ")
	fs.BoolVar(&useBase64, "base64", false, "base64 encode FUZZ values")
	fs.BoolVar(&useBase64, "B", false, "base64 encode FUZZ values")
	fs.BoolVar(&useURLEncode, "urlencode", false, "URL encode FUZZ values")
	fs.BoolVar(&useURLEncode, "E", false, "URL encode FUZZ values")
	setModuleUsage(fs, "fuzz", "FUZZ placeholder enumeration for URL, headers, auth, JSON body, or POST form data.", []string{
		"dirrunner fuzz -u 'https://example.com/FUZZ' -w wordlist/passwords.txt -B -E",
		"echo 'https://example.com/FUZZ' | dirrunner fuzz - -w wordlist/passwords.txt",
		"dirrunner fuzz -u https://example.com/login -p 'username=FUZZ&password=fake' -w username.txt",
		"dirrunner fuzz -u https://example.com/admin --user FUZZ --pass fake -w username.txt",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(true)}, helpGroup{Title: "HTTP flags", Flags: httpHelpFlags()}, helpGroup{Title: "Body flags", Flags: []helpFlag{
		{Names: "--json-body / -b", Use: "JSON", Desc: "JSON body containing FUZZ"},
		{Names: "--data / -p", Use: "FORM", Desc: "POST form data containing FUZZ"},
	}}, helpGroup{Title: "Encoder flags", Flags: []helpFlag{
		{Names: "--base64 / -B", Desc: "base64 encode FUZZ values"},
		{Names: "--urlencode / -E", Desc: "URL encode FUZZ values"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut
	var err error
	target, err = targetFromArgs(target, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	encoders := fuzzEncoders(useBase64, useURLEncode)
	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", target},
		{"Method", method},
		{"Wordlist", *common.wordlist},
		{"Status codes", codesRaw},
		{"JSON body", activeValue(bodyJSON != "")},
		{"POST data", activeValue(bodyForm != "")},
		{"Encoders", fuzzEncoderDisplay(encoders)},
	}...)
	showOptions("Fuzz Enumeration", common.commonFlags, rows)
	codes, err := parseCodes(codesRaw)
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	httpOpts := common.options()
	httpOpts.Method = method
	results, err := runner.RunFuzz(ctx, runner.FuzzOptions{
		TargetURL: target, Wordlist: *common.wordlist, Workers: *common.workers,
		Codes: codes, BodyJSON: bodyJSON, BodyForm: bodyForm, HTTP: httpOpts, Encoders: encoders,
	})
	return results, *common.jsonOut, *common.export, err
}

func fuzzEncoders(base64Enabled, urlencodeEnabled bool) []runner.FuzzEncoder {
	var encoders []runner.FuzzEncoder
	if base64Enabled {
		encoders = append(encoders, runner.FuzzEncoderBase64)
	}
	if urlencodeEnabled {
		encoders = append(encoders, runner.FuzzEncoderURLEncode)
	}
	return encoders
}

func fuzzEncoderDisplay(encoders []runner.FuzzEncoder) string {
	if len(encoders) == 0 {
		return "disabled"
	}
	labels := make([]string, 0, len(encoders))
	for _, encoder := range encoders {
		labels = append(labels, string(encoder))
	}
	return strings.Join(labels, ",")
}
