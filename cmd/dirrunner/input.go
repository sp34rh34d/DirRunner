package main

import (
	"fmt"
	"io"
	"os"
	"strings"
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
