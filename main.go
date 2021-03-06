package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
)

func main() {
	var help bool
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			help = true
			break
		}
	}

	if help {
		fmt.Println(
			strings.Join([]string{
				"Gostall build go executable to GOBIN with any name you want.",
				"",
				"Usage:",
				"    gostall PATH_TO_PACKAGE BINARY_NAME",
			}, "\n"),
		)
		return
	}

	if len(os.Args) != 3 {
		fatalf("gostall: Need two positional arguments: gostall [path] [name]")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	gobin, err := GetGOBIN(ctx)
	if err != nil {
		fatalf("gostall: failed to get GOBIN: %v", err)
	}

	path, name := os.Args[1], os.Args[2]

	cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(gobin, name), path)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin

	if err := cmd.Run(); err != nil {
		fatalf("gostall: error: %v", err)
	}
}

func GetGOBIN(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "GOBIN")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	output = bytes.TrimSpace(output)
	if len(output) == 0 {
		return "", fmt.Errorf("GOBIN not set")
	}

	return string(output), err
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
