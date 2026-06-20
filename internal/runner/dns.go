package runner

import (
	"context"
	"net"
	"strings"

	"dirrunner/internal/output"
	"dirrunner/internal/validate"
)

type DNSOptions struct {
	Domain   string
	Wordlist string
	Workers  int
}

func RunDNS(ctx context.Context, opts DNSOptions) ([]output.Result, error) {
	if !validate.Domain(opts.Domain) {
		return nil, ErrInvalidTarget("domain")
	}
	resolver := net.DefaultResolver
	return RunWorkers(ctx, opts.Wordlist, opts.Workers, func(ctx context.Context, word string) (output.Result, bool) {
		host := strings.Trim(strings.TrimSpace(word)+"."+opts.Domain, ".")
		ips, err := resolver.LookupHost(ctx, host)
		if err != nil || len(ips) == 0 {
			return output.Result{}, false
		}
		return output.Result{Type: "dns", Target: opts.Domain, Host: host, IP: ips[0]}, true
	})
}
