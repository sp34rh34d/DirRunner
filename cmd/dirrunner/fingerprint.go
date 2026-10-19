package main

import (
	"context"
	"flag"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runFingerprint(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("fingerprint", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 1, 15*time.Second)
	target := ""
	urlFile := ""
	fs.StringVar(&target, "url", "", "target URL")
	fs.StringVar(&target, "u", "", "target URL")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	setModuleUsage(fs, "fingerprint", "HTTP headers and technology fingerprint.", []string{
		"dirrunner fingerprint -u https://example.com",
		"echo https://example.com | dirrunner fingerprint -",
		"dirrunner fingerprint --url-file hosts.txt -j",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()})
	fs.Parse(args)
	output.StreamResults = false
	targets, err := targetsFromArgs(target, urlFile, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", targetsDisplay(targets)},
	}...)
	showOptions("Website Fingerprint", common.commonFlags, rows)
	results := runTargets(ctx, targets, func(t string) ([]output.Result, error) {
		return runner.RunFingerprint(ctx, runner.FingerprintOptions{TargetURL: t, HTTP: common.options()})
	})
	return results, *common.jsonOut, *common.export, nil
}
