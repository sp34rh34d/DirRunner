package main

import (
	"context"
	"flag"

	"dirrunner/internal/output"
	"dirrunner/internal/runner"
)

func runDNS(ctx context.Context, args []string) ([]output.Result, bool, string, error) {
	fs := flag.NewFlagSet("dns", flag.ExitOnError)
	common := addCommonFlags(fs, "wordlist/subdomains.txt", 10)
	target := ""
	fs.StringVar(&target, "domain", "", "target domain")
	fs.StringVar(&target, "d", "", "target domain")
	setModuleUsage(fs, "dns", "Subdomain DNS enumeration.", []string{
		"dirrunner dns -d example.com -w wordlist/subdomains.txt",
		"echo example.com | dirrunner dns -",
	}, helpGroup{Title: "Global flags", Flags: globalHelpFlags(true)}, helpGroup{Title: "DNS flags", Flags: []helpFlag{
		{Names: "--domain / -d", Use: "DOMAIN", Desc: "target domain"},
	}})
	fs.Parse(args)
	output.StreamResults = !*common.jsonOut
	var err error
	target, err = targetFromArgs(target, fs.Args())
	if err != nil {
		return nil, *common.jsonOut, *common.export, err
	}
	showOptions("DNS Enumeration", common, [][2]string{
		{"Target domain", target},
		{"Wordlist", *common.wordlist},
	})
	results, err := runner.RunDNS(ctx, runner.DNSOptions{Domain: target, Wordlist: *common.wordlist, Workers: *common.workers})
	return results, *common.jsonOut, *common.export, err
}
