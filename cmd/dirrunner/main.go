package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"dirrunner/internal/output"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var results []output.Result
	var err error
	var jsonOut bool
	var export string

	switch os.Args[1] {
	case "dns":
		results, jsonOut, export, err = runDNS(ctx, os.Args[2:])
	case "dir":
		results, jsonOut, export, err = runDir(ctx, os.Args[2:])
	case "vhost":
		results, jsonOut, export, err = runVHost(ctx, os.Args[2:])
	case "fuzz":
		results, jsonOut, export, err = runFuzz(ctx, os.Args[2:])
	case "fingerprint":
		results, jsonOut, export, err = runFingerprint(ctx, os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if jsonOut || !output.StreamResults {
		if err := output.PrintResults(os.Stdout, results, jsonOut); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}
	if export != "" {
		if err := output.WriteResults(export, results, jsonOut); err != nil {
			fmt.Fprintln(os.Stderr, "export error:", err)
			os.Exit(1)
		}
		output.Info("exported %d results to %s", len(results), export)
	}
}
