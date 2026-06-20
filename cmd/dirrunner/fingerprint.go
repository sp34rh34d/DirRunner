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
	fs.StringVar(&target, "url", "", "target URL")
	fs.StringVar(&target, "u", "", "target URL")
	setModuleUsage(fs, "fingerprint", "HTTP headers and technology fingerprint.", []string{
		"dirrunner fingerprint -u https://example.com",
		"echo https://example.com | dirrunner fingerprint -",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()})
	fs.Parse(args)
	output.StreamResults = false
	var err error
	target, err = targetFromArgs(target, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", target},
	}...)
	showOptions("Website Fingerprint", common.commonFlags, rows)
	results, err := runner.RunFingerprint(ctx, runner.FingerprintOptions{TargetURL: target, HTTP: common.options()})
	return results, *common.jsonOut, *common.export, err
}
