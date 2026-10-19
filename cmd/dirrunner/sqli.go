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

func runSQLi(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("sqli", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 20, 10*time.Second)
	target := ""
	urlFile := ""
	payloadWordlist := runner.DefaultSQLiPayloads
	headerWordlist := runner.DefaultSQLiHeaders
	errorWordlist := runner.DefaultSQLiErrors
	booleanWordlist := runner.DefaultSQLiBoolean
	sleepWordlist := runner.DefaultSQLiSleep
	encode := ""
	xorKey := "DirRunner"
	sleepSeconds := 3
	method := ""
	bodyForm := ""
	bodyJSON := ""
	csrfField := ""
	csrfURL := ""
	regx := ""
	fingerprint := false
	continueOnSuccess := false
	fs.StringVar(&target, "url", "", "target URL (include query parameters to test them)")
	fs.StringVar(&target, "u", "", "target URL (include query parameters to test them)")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	fs.StringVar(&method, "method", "", "HTTP method (default GET; POST when a body is set)")
	fs.StringVar(&method, "X", "", "HTTP method (default GET; POST when a body is set)")
	fs.StringVar(&bodyForm, "data", "", "form body to test, e.g. 'user=FUZZ&pass=x' (FUZZ marks the field, else all fields)")
	fs.StringVar(&bodyForm, "p", "", "form body to test (alias)")
	fs.StringVar(&bodyJSON, "json-body", "", "JSON body to test (FUZZ marks the field, else all string fields)")
	fs.StringVar(&bodyJSON, "b", "", "JSON body to test (alias)")
	fs.StringVar(&csrfField, "csrf-field", "", "body field holding an anti-CSRF token; a fresh token+cookie is fetched before each request")
	fs.StringVar(&csrfURL, "csrf-url", "", "page to GET the CSRF token from (default: the target URL)")
	fs.StringVar(&regx, "regx", "", "also report responses matching this regex/word, and show the match")
	fs.BoolVar(&fingerprint, "fingerprint", false, "identify the DBMS engine/version (MySQL/MariaDB, MSSQL, PostgreSQL, Oracle, SQLite, MongoDB, GraphQL) instead of the full scan")
	fs.StringVar(&payloadWordlist, "payload-wordlist", runner.DefaultSQLiPayloads, "SQLi payload wordlist")
	fs.StringVar(&headerWordlist, "header-wordlist", runner.DefaultSQLiHeaders, "header-name wordlist to test")
	fs.StringVar(&errorWordlist, "error-wordlist", runner.DefaultSQLiErrors, "SQL error-pattern wordlist (regex per line)")
	fs.StringVar(&booleanWordlist, "boolean-wordlist", runner.DefaultSQLiBoolean, "boolean TRUE/FALSE pair wordlist (TAB separated)")
	fs.StringVar(&sleepWordlist, "sleep-wordlist", runner.DefaultSQLiSleep, "time-based payload template wordlist (%d = seconds)")
	fs.StringVar(&encode, "encode", "", "chain these encodings in order (base64,urlencode,xor); only the encoded payload is sent")
	fs.StringVar(&encode, "payload-encoding", "", "alias for --encode")
	fs.StringVar(&xorKey, "xor-key", "DirRunner", "XOR key used with --encode xor")
	fs.IntVar(&sleepSeconds, "sleep", 3, "delay in seconds for time-based (SLEEP) detection; 0 disables it")
	fs.BoolVar(&continueOnSuccess, "continue-on-success", false, "keep scanning a target after a finding")
	setModuleUsage(fs, "sqli", "SQL-injection scanner for query parameters, request headers and POST bodies (form/JSON login forms).", []string{
		"dirrunner sqli -u 'https://example.com/item?id=1'",
		"dirrunner sqli -u https://example.com/login -p 'username=FUZZ&password=x'",
		"dirrunner sqli -u https://example.com/login -p 'username=admin&password=admin'  # tests both fields",
		"dirrunner sqli -u https://example.com/api/login -b '{\"user\":\"FUZZ\",\"pass\":\"x\"}'",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()}, helpGroup{Title: "SQLi flags", Flags: []helpFlag{
		{Names: "--url / -u", Use: "URL", Desc: "target URL with query parameters"},
		{Names: "--url-file", Use: "FILE", Desc: "file with target URLs"},
		{Names: "--method / -X", Use: "METHOD", Desc: "HTTP method (POST implied with a body)"},
		{Names: "--data / -p", Use: "FORM", Desc: "form body; FUZZ marks the field, else all fields"},
		{Names: "--json-body / -b", Use: "JSON", Desc: "JSON body; FUZZ marks the field, else all string fields"},
		{Names: "--csrf-field", Use: "NAME", Desc: "refresh this anti-CSRF token+cookie before each request"},
		{Names: "--csrf-url", Use: "URL", Desc: "page to read the CSRF token from (default: target)"},
		{Names: "--regx", Use: "PATTERN", Desc: "also report responses matching this regex/word"},
		{Names: "--fingerprint", Desc: "identify the DBMS engine/version instead of the full scan"},
		{Names: "--payload-wordlist / -w", Use: "FILE", Desc: "SQLi payload list"},
		{Names: "--header-wordlist", Use: "FILE", Desc: "header names to test"},
		{Names: "--error-wordlist", Use: "FILE", Desc: "SQL error regex patterns"},
		{Names: "--boolean-wordlist", Use: "FILE", Desc: "TRUE/FALSE pairs (TAB separated)"},
		{Names: "--sleep-wordlist", Use: "FILE", Desc: "time-based templates (%d seconds)"},
		{Names: "--encode", Use: "LIST", Desc: "base64,urlencode,xor"},
		{Names: "--xor-key", Use: "KEY", Desc: "key for --encode xor"},
		{Names: "--sleep", Use: "N", Desc: "time-based delay seconds (0 disables)"},
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
	// -w/--wordlist overrides the payload list, matching the other modules.
	if *common.wordlist != "" {
		payloadWordlist = *common.wordlist
	}

	bodyMode := "query/headers only"
	if bodyJSON != "" {
		bodyMode = "JSON body"
	} else if bodyForm != "" {
		bodyMode = "form body"
	}
	rows := append(httpOptionRows(common), [][2]string{
		{"Targets", fmt.Sprint(len(targets))},
		{"Method", sqliMethodDisplay(method, bodyForm != "" || bodyJSON != "")},
		{"Body", bodyMode},
		{"CSRF field", displayOrDisabled(csrfField)},
		{"Regx", displayOrDisabled(regx)},
		{"Mode", sqliModeDisplay(fingerprint)},
		{"Payload wordlist", payloadWordlist},
		{"Header wordlist", headerWordlist},
		{"Error wordlist", errorWordlist},
		{"Encodings", displayOrDisabled(strings.Join(encodings, ","))},
		{"Time-based sleep", sqliSleepDisplay(sleepSeconds)},
		{"Continue on success", fmt.Sprint(continueOnSuccess)},
	}...)
	showOptions("SQLi Detection", common.commonFlags, rows)

	results, err := runner.RunSQLi(ctx, runner.SQLiOptions{
		Targets:           targets,
		PayloadWordlist:   payloadWordlist,
		HeaderWordlist:    headerWordlist,
		ErrorWordlist:     errorWordlist,
		Encodings:         encodings,
		XorKey:            xorKey,
		BooleanWordlist:   booleanWordlist,
		SleepWordlist:     sleepWordlist,
		SleepSeconds:      sleepSeconds,
		Method:            method,
		BodyForm:          bodyForm,
		BodyJSON:          bodyJSON,
		CSRFField:         csrfField,
		CSRFURL:           csrfURL,
		Regx:              regx,
		Fingerprint:       fingerprint,
		ContinueOnSuccess: continueOnSuccess,
		Workers:           *common.workers,
		HTTP:              common.options(),
	})
	return results, *common.jsonOut, *common.export, err
}

func sqliMethodDisplay(method string, hasBody bool) string {
	if method != "" {
		return method
	}
	if hasBody {
		return "POST"
	}
	return "GET"
}

func sqliModeDisplay(fingerprint bool) string {
	if fingerprint {
		return "DBMS fingerprint"
	}
	return "injection scan"
}

func sqliSleepDisplay(seconds int) string {
	if seconds <= 0 {
		return "disabled"
	}
	return fmt.Sprintf("%ds", seconds)
}
