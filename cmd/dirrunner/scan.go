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

func runDir(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("dir", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 10, 15*time.Second)
	target := ""
	method := "GET"
	dirWordlist := "wordlist/directory.txt"
	fileWordlist := "wordlist/filename.txt"
	extRaw := ""
	codesRaw := "200,301,302"
	recursive := false
	depth := 1
	wildcardCheck := false
	skipWildcard := false
	fs.StringVar(&target, "url", "", "target base URL")
	fs.StringVar(&target, "u", "", "target base URL")
	fs.StringVar(&method, "method", "GET", "HTTP method")
	fs.StringVar(&method, "X", "GET", "HTTP method")
	fs.StringVar(&dirWordlist, "dir-wordlist", "wordlist/directory.txt", "directory wordlist path")
	fs.StringVar(&fileWordlist, "file-wordlist", "wordlist/filename.txt", "file-name wordlist path")
	fs.StringVar(&fileWordlist, "f", "wordlist/filename.txt", "file-name wordlist path")
	fs.StringVar(&extRaw, "ext", "", "comma separated file extensions; enables file enumeration")
	fs.StringVar(&extRaw, "e", "", "comma separated file extensions; enables file enumeration")
	fs.StringVar(&codesRaw, "codes", "200,301,302", "comma separated status codes to show")
	fs.StringVar(&codesRaw, "c", "200,301,302", "comma separated status codes to show")
	fs.BoolVar(&recursive, "recursive", false, "enumerate discovered directories recursively")
	fs.BoolVar(&recursive, "r", false, "enumerate discovered directories recursively")
	fs.IntVar(&depth, "depth", 1, "maximum recursion depth")
	fs.IntVar(&depth, "D", 1, "maximum recursion depth")
	fs.BoolVar(&wildcardCheck, "wildcard-check", false, "check missing-path wildcard responses before enumeration")
	fs.BoolVar(&wildcardCheck, "W", false, "check missing-path wildcard responses before enumeration")
	fs.BoolVar(&skipWildcard, "skip-wildcard-check", false, "legacy: skip missing-path status check")
	setModuleUsage(fs, "dir", "Directory enumeration. File checks activate only with --ext / -e.", []string{
		"dirrunner dir -u https://example.com -w wordlist/directory.txt -c 200,301,302",
		"echo https://example.com | dirrunner dir -",
		"dirrunner dir -u https://example.com -w wordlist/directory.txt -f wordlist/filename.txt -e php,txt",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(true)}, helpGroup{Title: "HTTP flags", Flags: httpHelpFlags()}, helpGroup{Title: "Directory and file flags", Flags: []helpFlag{
		{Names: "--dir-wordlist", Use: "FILE", Desc: "directory wordlist path"},
		{Names: "--file-wordlist / -f", Use: "FILE", Desc: "file-name wordlist path"},
		{Names: "--ext / -e", Use: "LIST", Desc: "file extensions; enables file enumeration"},
	}}, helpGroup{Title: "Recursion and wildcard flags", Flags: []helpFlag{
		{Names: "--recursive / -r", Desc: "enumerate discovered directories recursively"},
		{Names: "--depth / -D", Use: "N", Desc: "maximum recursion depth"},
		{Names: "--wildcard-check / -W", Desc: "check wildcard responses before enumeration"},
		{Names: "--skip-wildcard-check", Desc: "legacy alias to skip wildcard checks"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut
	var err error
	target, err = targetFromArgs(target, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}

	if *common.wordlist != "" {
		dirWordlist = *common.wordlist
	}
	extensions := splitCSV(extRaw)
	includeFiles := len(extensions) > 0

	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", target},
		{"Method", method},
		{"Status codes", codesRaw},
		{"Directories", "enabled"},
		{"Directory wordlist", dirWordlist},
		{"Files", activeValue(includeFiles)},
		{"File wordlist", fileWordlist},
		{"Extensions", extensionDisplay(extensions)},
		{"Recursive", fmt.Sprint(recursive)},
		{"Max depth", fmt.Sprint(depth)},
		{"Wildcard check", fmt.Sprint(wildcardCheck && !skipWildcard)},
	}...)
	showOptions("Dir Enumeration", common.commonFlags, rows)

	codes, err := parseCodes(codesRaw)
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	httpOpts := common.options()
	httpOpts.Method = method
	results, err := runner.RunScan(ctx, runner.ScanOptions{
		Target:            target,
		DirectoryWordlist: dirWordlist,
		FileWordlist:      fileWordlist,
		Workers:           *common.workers,
		Codes:             codes,
		Extensions:        extensions,
		HTTP:              httpOpts,
		Recursive:         recursive,
		MaxDepth:          depth,
		IncludeDirs:       true,
		IncludeFiles:      includeFiles,
		SkipWildcardCheck: !wildcardCheck || skipWildcard,
	})
	return results, *common.jsonOut, *common.export, err
}

func extensionDisplay(ext []string) string {
	if len(ext) == 0 {
		return "disabled"
	}
	return strings.Join(ext, ",")
}
