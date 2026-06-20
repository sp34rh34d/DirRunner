package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"dirrunner/internal/output"
)

type helpFlag struct {
	Names string
	Use   string
	Desc  string
}

type helpGroup struct {
	Title string
	Flags []helpFlag
}

func usage() {
	output.Banner(os.Stderr)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  dirrunner <module> [flags]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Modules:")
	fmt.Fprintln(os.Stderr, "  dns          Subdomain DNS enumeration")
	fmt.Fprintln(os.Stderr, "  dir          Directory enumeration; file checks activate with --ext/-e")
	fmt.Fprintln(os.Stderr, "  vhost        Virtual host enumeration through Host header")
	fmt.Fprintln(os.Stderr, "  fuzz         FUZZ placeholder enumeration")
	fmt.Fprintln(os.Stderr, "  fingerprint  HTTP headers and technology fingerprint")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  dirrunner dns -d example.com -w wordlist/subdomains.txt")
	fmt.Fprintln(os.Stderr, "  dirrunner dir -u https://example.com -w wordlist/directory.txt -c 200,301,302")
	fmt.Fprintln(os.Stderr, "  dirrunner dir -u https://example.com -w wordlist/directory.txt -f wordlist/filename.txt -e php,txt")
	fmt.Fprintln(os.Stderr, "  dirrunner vhost -u https://192.0.2.10 -d example.com")
	fmt.Fprintln(os.Stderr, "  dirrunner fuzz -u 'https://example.com/FUZZ' -w wordlist/passwords.txt -B -E")
	fmt.Fprintln(os.Stderr, "  dirrunner fingerprint -u https://example.com")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Global flags: all modules")
	fmt.Fprintln(os.Stderr, "  --threads / -t N       Concurrent workers, default 10")
	fmt.Fprintln(os.Stderr, "  --json / -j            Print JSON output")
	fmt.Fprintln(os.Stderr, "  --export / -o FILE     Write results to file")
	fmt.Fprintln(os.Stderr, "  --verbose / -v         Print detailed diagnostics")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Wordlist flags: dns, dir, vhost, fuzz")
	fmt.Fprintln(os.Stderr, "  --wordlist / -w FILE   Main wordlist")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "HTTP flags: dir, vhost, fuzz, fingerprint")
	fmt.Fprintln(os.Stderr, "  --url / -u URL                 Target URL")
	fmt.Fprintln(os.Stderr, "  --timeout / -T DURATION        Request timeout, e.g. 5s")
	fmt.Fprintln(os.Stderr, "  --user-agent / -A VALUE        User-Agent header")
	fmt.Fprintln(os.Stderr, "  --cookie / -C VALUE            Cookie header")
	fmt.Fprintln(os.Stderr, "  --header / -H 'Name: value'    Custom header, repeatable")
	fmt.Fprintln(os.Stderr, "  --username / --user / -U VALUE Basic auth username")
	fmt.Fprintln(os.Stderr, "  --password / --pass / -P VALUE Basic auth password")
	fmt.Fprintln(os.Stderr, "  --insecure / -k                Disable TLS certificate validation")
	fmt.Fprintln(os.Stderr, "  --follow-redirects / -L        Follow redirects")
	fmt.Fprintln(os.Stderr, "  --exclude-size / -S LIST       Hide exact response sizes")
	fmt.Fprintln(os.Stderr, "  --exclude-size-range / -R LIST Hide inclusive size ranges")
	fmt.Fprintln(os.Stderr, "  --tor / -Q                     Route HTTP requests through Tor")
	fmt.Fprintln(os.Stderr, "  --tor-proxy / -q HOST:PORT     Tor SOCKS5 proxy, default 127.0.0.1:9050")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Module-specific flags:")
	fmt.Fprintln(os.Stderr, "  dns:         --domain / -d DOMAIN")
	fmt.Fprintln(os.Stderr, "  dir:         --codes / -c LIST | --method / -X METHOD | --ext / -e LIST | --file-wordlist / -f FILE")
	fmt.Fprintln(os.Stderr, "               --recursive / -r | --depth / -D N | --wildcard-check / -W")
	fmt.Fprintln(os.Stderr, "  vhost:       --domain / -d DOMAIN")
	fmt.Fprintln(os.Stderr, "  fuzz:        --codes / -c LIST | --method / -X METHOD | --json-body / -b JSON | --data / -p FORM")
	fmt.Fprintln(os.Stderr, "               --base64 / -B | --urlencode / -E")
}

func setModuleUsage(fs *flag.FlagSet, module, summary string, examples []string, groups ...helpGroup) {
	fs.Usage = func() {
		output.Banner(os.Stderr)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Usage:\n  dirrunner %s [flags]\n\n", module)
		if summary != "" {
			fmt.Fprintln(os.Stderr, summary)
			fmt.Fprintln(os.Stderr)
		}
		if len(examples) > 0 {
			fmt.Fprintln(os.Stderr, "Examples:")
			for _, example := range examples {
				fmt.Fprintf(os.Stderr, "  %s\n", example)
			}
			fmt.Fprintln(os.Stderr)
		}
		for _, group := range groups {
			printHelpGroup(os.Stderr, group)
		}
	}
}

func printHelpGroup(w *os.File, group helpGroup) {
	if len(group.Flags) == 0 {
		return
	}
	fmt.Fprintf(w, "%s:\n", group.Title)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, flag := range group.Flags {
		usage := flag.Names
		if flag.Use != "" {
			usage += " " + flag.Use
		}
		fmt.Fprintf(tw, "  %s\t%s\n", usage, flag.Desc)
	}
	tw.Flush()
	fmt.Fprintln(w)
}

func globalHelpFlags(includeWordlist bool) []helpFlag {
	flags := []helpFlag{
		{Names: "--threads / -t", Use: "N", Desc: "concurrent workers, default 10"},
		{Names: "--json / -j", Desc: "print JSON output"},
		{Names: "--export / -o", Use: "FILE", Desc: "write results to a file"},
		{Names: "--verbose / -v", Desc: "print detailed diagnostics"},
	}
	if includeWordlist {
		flags = append([]helpFlag{{Names: "--wordlist / -w", Use: "FILE", Desc: "main wordlist path"}}, flags...)
	}
	return flags
}

func httpHelpFlags() []helpFlag {
	flags := []helpFlag{
		{Names: "--method / -X", Use: "METHOD", Desc: "HTTP method"},
		{Names: "--codes / -c", Use: "LIST", Desc: "status codes to show"},
	}
	return append(flags, httpBaseHelpFlags()...)
}

func httpBaseHelpFlags() []helpFlag {
	return []helpFlag{
		{Names: "--url / -u", Use: "URL", Desc: "target URL"},
		{Names: "--timeout / -T", Use: "DURATION", Desc: "request timeout, for example 5s"},
		{Names: "--user-agent / -A", Use: "VALUE", Desc: "User-Agent header"},
		{Names: "--cookie / -C", Use: "VALUE", Desc: "Cookie header"},
		{Names: "--header / -H", Use: "'Name: value'", Desc: "custom header, repeatable"},
		{Names: "--username / --user / -U", Use: "VALUE", Desc: "basic auth username"},
		{Names: "--password / --pass / -P", Use: "VALUE", Desc: "basic auth password"},
		{Names: "--insecure / -k", Desc: "disable TLS certificate validation"},
		{Names: "--follow-redirects / -L", Desc: "follow redirects"},
		{Names: "--exclude-size / -S", Use: "LIST", Desc: "hide exact response sizes"},
		{Names: "--exclude-size-range / -R", Use: "LIST", Desc: "hide inclusive size ranges"},
		{Names: "--tor / -Q", Desc: "route HTTP requests through Tor SOCKS5"},
		{Names: "--tor-proxy / -q", Use: "HOST:PORT", Desc: "Tor SOCKS5 proxy, default 127.0.0.1:9050"},
	}
}
