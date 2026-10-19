package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runCrawl(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("crawl", flag.ExitOnError)
	common := addHTTPFlags(fs, "", 20, 10*time.Second)
	target := ""
	urlFile := ""
	maxDepth := 3
	emails := false
	regex := ""
	allowExternal := false
	fs.StringVar(&target, "url", "", "target URL")
	fs.StringVar(&target, "u", "", "target URL")
	fs.StringVar(&urlFile, "url-file", "", "file with target URLs, one per line")
	fs.IntVar(&maxDepth, "max-depth", 3, "maximum crawl depth")
	fs.IntVar(&maxDepth, "D", 3, "maximum crawl depth")
	fs.BoolVar(&emails, "emails", false, "extract e-mail addresses from each page")
	fs.StringVar(&regex, "regx", "", "report every match of this regular expression")
	fs.StringVar(&regex, "regex", "", "alias for --regx")
	fs.BoolVar(&allowExternal, "allow-external", false, "follow links to other hosts (default: same host only)")
	setModuleUsage(fs, "crawl", "Web crawler: discovers page URLs plus JS/image assets. --emails and --regx add extraction over each page (WebRunner scraping + email-extractor + regx).", []string{
		"dirrunner crawl -u https://example.com --max-depth 2",
		"dirrunner crawl -u https://example.com --emails",
		"dirrunner crawl -u https://example.com --regx 'AKIA[0-9A-Z]{16}'",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(false)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()}, helpGroup{Title: "Crawl flags", Flags: []helpFlag{
		{Names: "--url / -u", Use: "URL", Desc: "target URL"},
		{Names: "--url-file", Use: "FILE", Desc: "file with target URLs"},
		{Names: "--max-depth / -D", Use: "N", Desc: "maximum crawl depth"},
		{Names: "--emails", Desc: "extract e-mail addresses"},
		{Names: "--regx", Use: "PATTERN", Desc: "report regex matches per page (alias: --regex)"},
		{Names: "--allow-external", Desc: "follow links to other hosts"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut

	targets, err := targetsFromArgs(target, urlFile, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}

	rows := append(httpOptionRows(common), [][2]string{
		{"Targets", fmt.Sprint(len(targets))},
		{"Max depth", fmt.Sprint(maxDepth)},
		{"Extract emails", activeValue(emails)},
		{"Regex", displayOrDisabled(regex)},
		{"Scope", crawlScope(allowExternal)},
	}...)
	showOptions("Web Crawler", common.commonFlags, rows)

	results, err := runner.RunCrawl(ctx, runner.CrawlOptions{
		Targets:  targets,
		MaxDepth: maxDepth,
		Emails:   emails,
		Regex:    regex,
		SameHost: !allowExternal,
		Workers:  *common.workers,
		HTTP:     common.options(),
	})
	return results, *common.jsonOut, *common.export, err
}

func crawlScope(allowExternal bool) string {
	if allowExternal {
		return "all hosts"
	}
	return "same host"
}
