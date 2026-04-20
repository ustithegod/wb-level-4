package main

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ustithegod/wb-level-4/quorum-cut/internal/node"
)

func TestLocalMatchesSystemCut(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("cut"); err != nil {
		t.Skip("system cut not available")
	}

	file := writeSampleFile(t)
	args := []string{"local", "-d", ":", "-f", "1,3", file}

	gotStdout, gotStderr, gotErr := runMainForTest(t, args, nil)
	wantStdout, wantStderr, wantErr := runSystemCut(t, []string{"-d", ":", "-f", "1,3", file}, nil)

	compareRunResult(t, gotStdout, gotStderr, gotErr, wantStdout, wantStderr, wantErr)
}

func TestDistributedAllMatchesSystemCut(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("cut"); err != nil {
		t.Skip("system cut not available")
	}

	srv1 := httptest.NewServer(node.NewServer("n1").Handler())
	defer srv1.Close()
	srv2 := httptest.NewServer(node.NewServer("n2").Handler())
	defer srv2.Close()
	srv3 := httptest.NewServer(node.NewServer("n3").Handler())
	defer srv3.Close()

	file := writeSampleFile(t)
	nodes := strings.Join([]string{
		strings.TrimPrefix(srv1.URL, "http://"),
		strings.TrimPrefix(srv2.URL, "http://"),
		strings.TrimPrefix(srv3.URL, "http://"),
	}, ",")

	args := []string{
		"search",
		"--nodes", nodes,
		"--consistency", "all",
		"--chunk-size", "1",
		"-d", ":",
		"-f", "2",
		file,
	}

	gotStdout, gotStderr, gotErr := runMainForTest(t, args, nil)
	wantStdout, wantStderr, wantErr := runSystemCut(t, []string{"-d", ":", "-f", "2", file}, nil)

	compareRunResult(t, gotStdout, gotStderr, gotErr, wantStdout, wantStderr, wantErr)
}

func TestDistributedAllMatchesSystemCutFromStdin(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("cut"); err != nil {
		t.Skip("system cut not available")
	}

	srv1 := httptest.NewServer(node.NewServer("n1").Handler())
	defer srv1.Close()
	srv2 := httptest.NewServer(node.NewServer("n2").Handler())
	defer srv2.Close()

	nodes := strings.Join([]string{
		strings.TrimPrefix(srv1.URL, "http://"),
		strings.TrimPrefix(srv2.URL, "http://"),
	}, ",")

	stdinData := []byte("alpha,beta,gamma\nx,y,z\n")
	args := []string{
		"search",
		"--nodes", nodes,
		"--consistency", "all",
		"-d", ",",
		"-f", "1,3",
	}

	gotStdout, gotStderr, gotErr := runMainForTest(t, args, bytes.NewReader(stdinData))
	wantStdout, wantStderr, wantErr := runSystemCut(t, []string{"-d", ",", "-f", "1,3"}, stdinData)

	compareRunResult(t, gotStdout, gotStderr, gotErr, wantStdout, wantStderr, wantErr)
}

func runMainForTest(t *testing.T, args []string, stdin io.Reader) (string, string, error) {
	t.Helper()

	if stdin == nil {
		stdin = bytes.NewReader(nil)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(args, stdin, &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

func runSystemCut(t *testing.T, args []string, stdin []byte) (string, string, error) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "cut", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

func compareRunResult(t *testing.T, gotStdout, gotStderr string, gotErr error, wantStdout, wantStderr string, wantErr error) {
	t.Helper()

	if gotStdout != wantStdout {
		t.Fatalf("stdout mismatch\n got: %q\nwant: %q", gotStdout, wantStdout)
	}
	if gotStderr != wantStderr {
		t.Fatalf("stderr mismatch\n got: %q\nwant: %q", gotStderr, wantStderr)
	}
	if (gotErr != nil) != (wantErr != nil) {
		t.Fatalf("error presence mismatch\n got: %v\nwant: %v", gotErr, wantErr)
	}
}

func writeSampleFile(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	content := "left:middle:right\n1:2:3\nplain:field:value\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}
	return path
}
