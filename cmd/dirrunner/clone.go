package main

import (
	"context"
	"flag"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runClone(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("clone", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 10, 15*time.Second)
	target := ""
	urlFile := ""
	name := ""
	fs.StringVar(&target, "url", "", "target URL to clone")
	fs.StringVar(&target, "u", "", "target URL to clone")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	fs.StringVar(&name, "name", "", "project directory name (default: generated)")
	fs.StringVar(&name, "n", "", "project directory name (default: generated)")
	setModuleUsage(fs, "clone", "Download a site breadth-first, following same-domain links and assets, saving every resource to disk.", []string{
		"dirrunner clone -u https://example.com -n example-copy",
		"echo https://example.com | dirrunner clone -",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()}, helpGroup{Title: "Clone flags", Flags: []helpFlag{
		{Names: "--url / -u", Use: "URL", Desc: "target URL to clone"},
		{Names: "--name / -n", Use: "NAME", Desc: "project directory name"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut

	targets, err := targetsFromArgs(target, urlFile, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}

	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", targetsDisplay(targets)},
		{"Project name", displayOrDisabled(name)},
	}...)
	showOptions("Web Cloner", common.commonFlags, rows)

	results := runTargets(ctx, targets, func(t string) ([]output.Result, error) {
		return runner.RunClone(ctx, runner.CloneOptions{
			Target:      t,
			ProjectName: name,
			Workers:     *common.workers,
			HTTP:        common.options(),
		})
	})
	return results, *common.jsonOut, *common.export, nil
}
