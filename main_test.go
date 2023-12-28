package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var coverDir string

func init() {
	for i := len(os.Args) - 1; i > -1; i-- {
		_, dir, ok := strings.Cut(os.Args[i], "-test.gocoverdir=")
		if !ok {
			continue
		}

		coverDir = Must2(filepath.Abs(dir))
		Must(os.RemoveAll(coverDir))
		Must(os.MkdirAll(coverDir, 0o755))
		return
	}
}

func TestMain(t *testing.T) {
	file, err := os.CreateTemp("", "")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	args := []string{"build", "-o", file.Name()}
	if coverDir != "" {
		args = append(args, "-cover", "-coverpkg=github.com/davidmdm/gostall")
	}
	args = append(args, ".")

	build := exec.Command("go", args...)
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	require.NoError(t, build.Run())

	gobin, err := filepath.Abs(outputGobin)
	require.NoError(t, err)

	gostall := func(path, name string) *exec.Cmd {
		cmd := exec.Command(file.Name(), path, name)
		cmd.Dir = outputDir
		cmd.Env = append(os.Environ(), "GOBIN="+gobin, "GOCOVERDIR="+coverDir)
		return cmd
	}

	cases := []struct {
		Name        string
		Path        string
		Out         string
		ExpectedOut string
	}{
		{
			Name:        "local path with name",
			Path:        "..",
			Out:         "test",
			ExpectedOut: filepath.Join(gobin, "test"),
		},
		{
			Name:        "local path with local out",
			Path:        "..",
			Out:         "./test",
			ExpectedOut: filepath.Join(outputDir, "test"),
		},
		{
			Name:        "remote path with name",
			Path:        "github.com/matryer/moq@latest",
			Out:         "test",
			ExpectedOut: filepath.Join(gobin, "test"),
		},
		{
			Name:        "remote path with local out",
			Path:        "github.com/matryer/moq@latest",
			Out:         "./test",
			ExpectedOut: filepath.Join(outputDir, "test"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ResetTestOutput(t)

			require.NoError(t, gostall(tc.Path, tc.Out).Run())

			_, err := os.Stat(tc.ExpectedOut)
			require.NoError(t, err)
		})
	}

	t.Run("no args passed", func(t *testing.T) {
		output, err := exec.Command(file.Name()).CombinedOutput()
		require.EqualError(t, err, "exit status 1")
		require.Contains(t, string(output), "need two positional arguments: [path] [name]\n")
	})
}

var (
	outputDir   = "./test_output"
	outputGobin = filepath.Join(outputDir, "bin")
)

func ResetTestOutput(t *testing.T) {
	t.Helper()
	require.NoError(t, os.RemoveAll(outputDir))
	require.NoError(t, os.MkdirAll(outputGobin, 0o755))
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Must2[T any](value T, err error) T {
	Must(err)
	return value
}
