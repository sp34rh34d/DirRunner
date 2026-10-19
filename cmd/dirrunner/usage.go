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
	fmt.Fprintln(os.Stderr, "  traversal    Path-traversal / local file disclosure scanner")
	fmt.Fprintln(os.Stderr, "  bypass       401/403 access-control bypass probe")
	fmt.Fprintln(os.Stderr, "  sqli         SQL-injection indicator scanner")
	fmt.Fprintln(os.Stderr, "  crawl        Web crawler (URLs/JS/images, e-mails, regex)")
	fmt.Fprintln(os.Stderr, "  clone        Download a site to disk")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  dirrunner dns -d example.com -w wordlist/subdomains.txt")
	fmt.Fprintln(os.Stderr, "  dirrunner dir -u https://example.com -w wordlist/directory.txt -c 200,301,302")
	fmt.Fprintln(os.Stderr, "  dirrunner dir -u https://example.com -w wordlist/directory.txt -f wordlist/filename.txt -e php,txt")
	fmt.Fprintln(os.Stderr, "  dirrunner vhost -u https://192.0.2.10 -d example.com")
	fmt.Fprintln(os.Stderr, "  dirrunner fuzz -u 'https://example.com/FUZZ' -w wordlist/passwords.txt --encode base64,urlencode")
	fmt.Fprintln(os.Stderr, "  dirrunner fingerprint -u https://example.com")
	fmt.Fprintln(os.Stderr, "  dirrunner traversal -u 'https://example.com/?page=FUZZ' --os all")
	fmt.Fprintln(os.Stderr, "  dirrunner bypass -u https://example.com/admin")
	fmt.Fprintln(os.Stderr, "  dirrunner sqli -u 'https://example.com/item?id=1'")
	fmt.Fprintln(os.Stderr, "  dirrunner crawl -u https://example.com --emails")
	fmt.Fprintln(os.Stderr, "  dirrunner clone -u https://example.com -n example-copy")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Global flags: all modules")
	fmt.Fprintln(os.Stderr, "  --threads / -t N       Concurrent workers, default 40")
	fmt.Fprintln(os.Stderr, "  --json / -j            Print JSON output")
	fmt.Fprintln(os.Stderr, "  --export / -o FILE     Write results to file")
	fmt.Fprintln(os.Stderr, "  --verbose / -v         Print detailed diagnostics")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Wordlist flags: dns, dir, vhost, fuzz, traversal, bypass, sqli")
	fmt.Fprintln(os.Stderr, "  --wordlist / -w FILE   Main wordlist (payloads for sqli, headers for bypass)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "HTTP flags: dir, vhost, fuzz, fingerprint, traversal, bypass, sqli, crawl, clone")
	fmt.Fprintln(os.Stderr, "  --url-file FILE   Scan many target URLs from a file (every URL module; also piped via -)")
	fmt.Fprintln(os.Stderr, "  --url / -u URL                 Target URL")
	fmt.Fprintln(os.Stderr, "  --timeout / -T DURATION        Request timeout, e.g. 5s")
	fmt.Fprintln(os.Stderr, "  --user-agent / -A VALUE        User-Agent header")
	fmt.Fprintln(os.Stderr, "  --random-agent / -a            Random User-Agent per request")
	fmt.Fprintln(os.Stderr, "  --cookie / -C VALUE            Cookie header")
	fmt.Fprintln(os.Stderr, "  --header / -H 'Name: value'    Custom header, repeatable")
	fmt.Fprintln(os.Stderr, "  --username / --user / -U VALUE Basic auth username")
	fmt.Fprintln(os.Stderr, "  --password / --pass / -P VALUE Basic auth password")
	fmt.Fprintln(os.Stderr, "  --insecure / -k                Disable TLS certificate validation")
	fmt.Fprintln(os.Stderr, "  --follow-redirects / -L        Follow redirects")
	fmt.Fprintln(os.Stderr, "  --exclude-size / -S LIST       Hide exact response sizes")
	fmt.Fprintln(os.Stderr, "  --exclude-size-range / -R LIST Hide inclusive size ranges")
	fmt.Fprintln(os.Stderr, "  --proxy / -x URL               http/https/socks5/socks5h proxy (Tor: socks5h://127.0.0.1:9050)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Module-specific flags:")
	fmt.Fprintln(os.Stderr, "  dns:         --domain / -d DOMAIN")
	fmt.Fprintln(os.Stderr, "  dir:         --codes / -c LIST | --method / -X METHOD | --ext / -e LIST | --file-wordlist / -f FILE")
	fmt.Fprintln(os.Stderr, "               --recursive / -r | --depth / -D N | --wildcard-check / -W")
	fmt.Fprintln(os.Stderr, "  vhost:       --domain / -d DOMAIN")
	fmt.Fprintln(os.Stderr, "  fuzz:        --codes / -c LIST | --method / -X METHOD | --json-body / -b JSON | --data / -p FORM")
	fmt.Fprintln(os.Stderr, "               --csrf-field NAME | --csrf-url URL | --encode base64,urlencode,xor | --xor-key KEY | -B | -E")
	fmt.Fprintln(os.Stderr, "  traversal:   --url-file FILE | --files-wordlist FILE | --min-depth / --max-depth N")
	fmt.Fprintln(os.Stderr, "               --signature-wordlist FILE | --regx PATTERN | --encode base64,urlencode,xor")
	fmt.Fprintln(os.Stderr, "               --custom-path/-string/-file LIST | --continue-on-success")
	fmt.Fprintln(os.Stderr, "  bypass:      --url-file FILE | --header-wordlist FILE | --path-wordlist FILE")
	fmt.Fprintln(os.Stderr, "               --bypass-header 'Name: value' | --bypass-path SRC | --continue-on-success")
	fmt.Fprintln(os.Stderr, "  sqli:        --url-file FILE | --method / -X METHOD | --data / -p FORM | --json-body / -b JSON")
	fmt.Fprintln(os.Stderr, "               --csrf-field NAME | --csrf-url URL | --regx PATTERN | --fingerprint")
	fmt.Fprintln(os.Stderr, "               --payload/-header/-error/-boolean/-sleep-wordlist FILE")
	fmt.Fprintln(os.Stderr, "               --encode base64,urlencode,xor | --sleep N | --continue-on-success")
	fmt.Fprintln(os.Stderr, "  crawl:       --url-file FILE | --max-depth / -D N | --emails | --regx PATTERN | --allow-external")
	fmt.Fprintln(os.Stderr, "  clone:       --name / -n NAME")
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
		{Names: "--threads / -t", Use: "N", Desc: "concurrent workers, default 40"},
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
		{Names: "--random-agent / -a", Desc: "random User-Agent per request"},
		{Names: "--cookie / -C", Use: "VALUE", Desc: "Cookie header"},
		{Names: "--header / -H", Use: "'Name: value'", Desc: "custom header, repeatable"},
		{Names: "--username / --user / -U", Use: "VALUE", Desc: "basic auth username"},
		{Names: "--password / --pass / -P", Use: "VALUE", Desc: "basic auth password"},
		{Names: "--insecure / -k", Desc: "disable TLS certificate validation"},
		{Names: "--follow-redirects / -L", Desc: "follow redirects"},
		{Names: "--exclude-size / -S", Use: "LIST", Desc: "hide exact response sizes"},
		{Names: "--exclude-size-range / -R", Use: "LIST", Desc: "hide inclusive size ranges"},
		{Names: "--proxy / -x", Use: "URL", Desc: "http/https/socks5/socks5h proxy (Tor: socks5h://127.0.0.1:9050)"},
	}
}
