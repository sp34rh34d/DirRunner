package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runBypass(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("bypass", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 20, 10*time.Second)
	target := ""
	urlFile := ""
	headerWordlist := runner.DefaultBypassHeaders
	pathWordlist := runner.DefaultBypassPaths
	var bypassHeaders arrayFlag
	bypassPath := ""
	continueOnSuccess := false
	fs.StringVar(&target, "url", "", "target URL")
	fs.StringVar(&target, "u", "", "target URL")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	fs.StringVar(&headerWordlist, "header-wordlist", runner.DefaultBypassHeaders, "base bypass header wordlist")
	fs.StringVar(&pathWordlist, "path-wordlist", runner.DefaultBypassPaths, "base bypass path-suffix wordlist")
	fs.Var(&bypassHeaders, "bypass-header", "extra bypass header, repeatable: Name: value")
	fs.StringVar(&bypassPath, "bypass-path", "", "extra path, or file/dir with paths, to test")
	fs.BoolVar(&continueOnSuccess, "continue-on-success", false, "keep scanning a target after a finding")
	setModuleUsage(fs, "bypass", "401/403 access-control bypass probe (header tricks and path mutations).", []string{
		"dirrunner bypass -u https://example.com/admin",
		"dirrunner bypass -u https://example.com/admin --bypass-header 'X-Real-IP: 127.0.0.1'",
		"dirrunner bypass --url-file protected.txt --continue-on-success",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()}, helpGroup{Title: "Bypass flags", Flags: []helpFlag{
		{Names: "--url / -u", Use: "URL", Desc: "target URL"},
		{Names: "--url-file", Use: "FILE", Desc: "file with target URLs"},
		{Names: "--header-wordlist / -w", Use: "FILE", Desc: "base bypass header list"},
		{Names: "--path-wordlist", Use: "FILE", Desc: "base path-suffix list"},
		{Names: "--bypass-header", Use: "'Name: value'", Desc: "extra header, repeatable"},
		{Names: "--bypass-path", Use: "SRC", Desc: "extra path, or file/dir of paths"},
		{Names: "--continue-on-success", Desc: "keep scanning after a finding"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut

	targets, err := targetsFromArgs(target, urlFile, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	// -w/--wordlist overrides the base header list, matching the other modules.
	if *common.wordlist != "" {
		headerWordlist = *common.wordlist
	}

	rows := append(httpOptionRows(common), [][2]string{
		{"Targets", fmt.Sprint(len(targets))},
		{"Header wordlist", headerWordlist},
		{"Path wordlist", pathWordlist},
		{"Extra headers", fmt.Sprint(len(bypassHeaders))},
		{"Extra path source", displayOrDisabled(bypassPath)},
		{"Continue on success", fmt.Sprint(continueOnSuccess)},
	}...)
	showOptions("403 Bypass", common.commonFlags, rows)

	results, err := runner.RunBypass(ctx, runner.BypassOptions{
		Targets:           targets,
		HeaderWordlist:    headerWordlist,
		PathWordlist:      pathWordlist,
		CustomHeaders:     bypassHeaders,
		CustomPathSource:  bypassPath,
		ContinueOnSuccess: continueOnSuccess,
		Workers:           *common.workers,
		HTTP:              common.options(),
	})
	return results, *common.jsonOut, *common.export, err
}
