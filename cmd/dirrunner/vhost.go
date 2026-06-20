package main

import (
	"context"
	"flag"
	"time"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runVHost(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("vhost", flag.ExitOnError)
	common := addHTTPFlags(fs, "wordlist/subdomains.txt", 10, 2*time.Second)
	targetURL := ""
	domain := ""
	fs.StringVar(&targetURL, "url", "", "target URL or IP")
	fs.StringVar(&targetURL, "u", "", "target URL or IP")
	fs.StringVar(&domain, "domain", "", "domain used to build Host headers")
	fs.StringVar(&domain, "d", "", "domain used to build Host headers")
	setModuleUsage(fs, "vhost", "Virtual host enumeration through the Host header.", []string{
		"dirrunner vhost -u https://192.0.2.10 -d example.com -w wordlist/subdomains.txt",
		"echo https://192.0.2.10 | dirrunner vhost - -d example.com -w wordlist/subdomains.txt",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(true)}, helpGroup{Title: "HTTP flags", Flags: httpBaseHelpFlags()}, helpGroup{Title: "VHost flags", Flags: []helpFlag{
		{Names: "--domain / -d", Use: "DOMAIN", Desc: "domain used to build Host headers"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut
	var err error
	targetURL, err = targetFromArgs(targetURL, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	rows := append(httpOptionRows(common), [][2]string{
		{"Target URL", targetURL},
		{"Target domain", domain},
		{"Wordlist", *common.wordlist},
	}...)
	showOptions("Virtual Host Enumeration", common.commonFlags, rows)
	results, err := runner.RunVHost(ctx, runner.VHostOptions{
		TargetURL: targetURL, Domain: domain, Wordlist: *common.wordlist,
		Workers: *common.workers, HTTP: common.options(),
	})
	return results, *common.jsonOut, *common.export, err
}
