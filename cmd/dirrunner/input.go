package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"dirrunner/internal/runner"
)

func targetFromArgs(current string, args []string) (string, error) {
	if strings.TrimSpace(current) != "" {
		return strings.TrimSpace(current), nil
	}
	if len(args) == 0 {
		return "", nil
	}
	raw := strings.TrimSpace(args[0])
	if raw != "-" {
		return raw, nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		target := strings.TrimSpace(line)
		if target != "" {
			return target, nil
		}
	}
	return "", fmt.Errorf("stdin did not provide a target URL")
}

// targetsFromArgs resolves one or more target URLs for the web modules, from
// (in order) --url-file, --url/-u, a positional "-" that reads URLs from stdin,
// or a single positional URL. Only valid http(s) URLs are kept.
func targetsFromArgs(current, urlFile string, args []string) ([]string, error) {
	var raw []string
	switch {
	case strings.TrimSpace(urlFile) != "":
		data, err := os.ReadFile(urlFile)
		if err != nil {
			return nil, err
		}
		raw = strings.Split(string(data), "\n")
	case strings.TrimSpace(current) != "":
		raw = []string{current}
	case len(args) > 0 && strings.TrimSpace(args[0]) == "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		raw = strings.Split(string(data), "\n")
	case len(args) > 0:
		raw = []string{args[0]}
	default:
		return nil, fmt.Errorf("no target provided; use --url/-u, --url-file, or pipe URLs with -")
	}

	var targets []string
	skipped := 0
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !runner.ValidURL(line) {
			skipped++
			continue
		}
		targets = append(targets, line)
	}
	if skipped > 0 {
		fmt.Fprintf(os.Stderr, "[!] skipped %d invalid (non http/https) target URL(s)\n", skipped)
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid http(s) target URLs found")
	}
	return targets, nil
}
