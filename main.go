package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/davidmdm/xcontext"
)

var usage = strings.Join([]string{
	"gostall",
	"",
	"build go executable to GOBIN with any name you want.",
	"",
	"Usage:",
	"    gostall PATH_TO_PACKAGE BINARY_NAME",
}, "\n")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = func() { fmt.Fprintln(os.Stderr, usage) }
	flag.Parse()

	if len(flag.Args()) != 2 {
		return errors.New("gostall: Need two positional arguments: gostall [path] [name]")
	}

	ctx, cancel := xcontext.WithSignalCancelation(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	gobin, err := GetGOBIN(ctx)
	if err != nil {
		return fmt.Errorf("gostall: failed to get GOBIN: %v", err)
	}

	path, name := os.Args[1], os.Args[2]

	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(gobin, name), path)
	build.Stdout, build.Stderr, build.Stdin = os.Stdout, os.Stderr, os.Stdin

	if err := build.Run(); err != nil {
		return fmt.Errorf("gostall: error: %v", err)
	}

	return nil
}

func GetGOBIN(ctx context.Context) (string, error) {
	output, err := exec.CommandContext(ctx, "go", "env", "GOBIN").Output()
	if err != nil {
		return "", err
	}

	output = bytes.TrimSpace(output)
	if len(output) == 0 {
		return "", fmt.Errorf("GOBIN not set")
	}

	return string(output), err
}
