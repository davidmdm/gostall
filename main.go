package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/davidmdm/x/xcontext"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//go:embed usage.txt
var usage string

func run() error {
	flag.Usage = func() { fmt.Fprintln(os.Stderr, usage) }
	flag.Parse()

	if len(flag.Args()) != 2 {
		return errors.New("gostall: Need two positional arguments: gostall [path] [name]")
	}

	ctx, cancel := xcontext.WithSignalCancelation(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	gobin, err := GetGoVar(ctx, "GOBIN")
	if err != nil {
		return fmt.Errorf("gostall: failed to get GOBIN: %v", err)
	}

	path, name := flag.Arg(0), flag.Arg(1)

	build := func() BuildFunc {
		if path == "." || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "/") {
			return buildLocalPath
		}
		return buildRemotePath
	}()

	return build(ctx, gobin, path, name)
}

type BuildFunc = func(ctx context.Context, gobin, path, name string) error

func buildLocalPath(ctx context.Context, gobin, path, name string) error {
	build := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(gobin, name), path)
	build.Stdout, build.Stderr, build.Stdin = os.Stdout, os.Stderr, os.Stdin
	return build.Run()
}

func buildRemotePath(ctx context.Context, gobin, path, name string) error {
	temp, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("failed to create temporary module: %w", err)
	}
	defer os.RemoveAll(temp)

	cmd := func(ctx context.Context, name string, args ...string) error {
		c := exec.CommandContext(ctx, name, args...)
		c.Dir = temp
		c.Stdout, c.Stderr, c.Stdin = os.Stdout, os.Stderr, os.Stdin
		return c.Run()
	}

	if err := cmd(ctx, "go", "mod", "init", "builder"); err != nil {
		return fmt.Errorf("failed to init temporary builder module: %w", err)
	}

	if err := cmd(ctx, "go", "get", path); err != nil {
		return fmt.Errorf("failed to get %s: %w", path, err)
	}

	base, _, _ := strings.Cut(path, "@")

	if err := cmd(ctx, "go", "build", "-o", filepath.Join(gobin, name), base); err != nil {
		return fmt.Errorf("failed to install %s: %w", base, err)
	}

	return nil
}

func GetGoVar(ctx context.Context, name string) (string, error) {
	if value := os.Getenv(name); value != "" {
		return value, nil
	}

	output, err := exec.CommandContext(ctx, "go", "env", name).Output()
	if err != nil {
		return "", err
	}

	output = bytes.TrimSpace(output)
	if len(output) == 0 {
		return "", fmt.Errorf("%s not set", name)
	}

	return string(output), err
}
