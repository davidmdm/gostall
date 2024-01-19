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
	"slices"
	"strings"
	"syscall"

	"github.com/davidmdm/x/xcontext"
)

var (
	Version    = "v0.0.7"
	binaryName = os.Args[0]
)

//go:embed usage.txt
var usage string

func init() {
	usage = fmt.Sprintf(usage, binaryName, binaryName)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", binaryName, err)
		os.Exit(1)
	}
}

func run() error {
	args, subargs := cut(os.Args[1:], "--")

	flag.CommandLine.Usage = func() { fmt.Fprintln(os.Stderr, usage) }
	flag.CommandLine.Parse(args)

	if flag.Arg(0) == "version" {
		fmt.Println(Version)
		return nil
	}

	if len(flag.Args()) < 2 {
		return errors.New("need two positional arguments: [path] [name]")
	}

	ctx, cancel := xcontext.WithSignalCancelation(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	path, name := flag.Arg(0), flag.Arg(1)

	outputFile, err := func() (string, error) {
		segments := strings.Split(name, string([]byte{os.PathSeparator}))
		if len(segments) != 1 {
			return filepath.Abs(name)
		}

		gobin, err := GetGoVar(ctx, "GOBIN")
		if err != nil {
			return "", fmt.Errorf("failed to get GOBIN: %v", err)
		}

		return filepath.Abs(filepath.Join(gobin, name))
	}()
	if err != nil {
		return fmt.Errorf("failed to determine outputfile for binary: %w", err)
	}

	build := func() BuildFunc {
		if _, err := os.Stat(path); err == nil {
			return buildLocalPath
		}
		return buildRemotePath
	}()

	return build(ctx, path, outputFile, subargs)
}

type BuildFunc = func(ctx context.Context, path, out string, args []string) error

func buildLocalPath(ctx context.Context, path, out string, args []string) error {
	args = append([]string{"build"}, args...)
	args = append(args, "-o", out, path)

	build := exec.CommandContext(ctx, "go", args...)
	build.Stdout, build.Stderr, build.Stdin = os.Stdout, os.Stderr, os.Stdin
	return build.Run()
}

func buildRemotePath(ctx context.Context, path, out string, args []string) error {
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

	args = append([]string{"build"}, args...)
	args = append(args, "-o", out, base)

	if err := cmd(ctx, "go", args...); err != nil {
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

func cut(args []string, sep string) (before, after []string) {
	idx := slices.Index(args, sep)
	if idx == -1 {
		return args, nil
	}
	return args[:idx], args[idx+1:]
}
