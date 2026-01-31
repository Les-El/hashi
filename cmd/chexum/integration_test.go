package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/testutil"
)

func TestCLI_Help(t *testing.T) {
	cmd := exec.Command(binaryName, "--help")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("chexum --help failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "USAGE") {
		t.Errorf("expected help output to contain USAGE, got: %s", output)
	}
}

func TestCLI_Version(t *testing.T) {
	cmd := exec.Command(binaryName, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("chexum --version failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "chexum version") {
		t.Errorf("expected version output to contain 'chexum version', got: %s", output)
	}
}

func TestCLI_HashFile(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	content := "hello integration test"
	filePath := testutil.CreateFile(t, tmpDir, "test.txt", content)

	cmd := exec.Command(binaryName, filePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("chexum %s failed: %v", filePath, err)
	}

	output := stdout.String()
	if !strings.Contains(output, "test.txt") {
		t.Errorf("expected output to contain filename, got: %s", output)
	}
}

func TestCLI_JSONOutput(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	testutil.CreateFile(t, tmpDir, "f1.txt", "content1")
	testutil.CreateFile(t, tmpDir, "f2.txt", "content1") // Same content

	cmd := exec.Command(binaryName, "--json", filepath.Join(tmpDir, "f1.txt"), filepath.Join(tmpDir, "f2.txt"))
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("chexum --json failed: %v", err)
	}

	var res struct {
		Processed   int `json:"processed"`
		MatchGroups []interface{} `json:"match_groups"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, stdout.String())
	}

	if res.Processed != 2 {
		t.Errorf("expected 2 processed files, got %d", res.Processed)
	}
	if len(res.MatchGroups) != 1 {
		t.Errorf("expected 1 match group, got %d", len(res.MatchGroups))
	}
}

func TestCLI_Recursive(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	testutil.CreateFile(t, subDir, "f1.txt", "c1")

	cmd := exec.Command(binaryName, "-r", tmpDir)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("chexum -r failed: %v", err)
	}

	if !strings.Contains(stdout.String(), "f1.txt") {
		t.Errorf("expected output to contain sub/f1.txt, got: %s", stdout.String())
	}
}

func TestCLI_Algorithm(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	filePath := testutil.CreateFile(t, tmpDir, "test.txt", "hello")

	cmd := exec.Command(binaryName, "-a", "md5", filePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		t.Fatalf("chexum -a md5 failed: %v", err)
	}

	// MD5 of "hello" is 5d41402abc4b2a76b9719d911017c592
	if !strings.Contains(stdout.String(), "5d41402abc4b2a76b9719d911017c592") {
		t.Errorf("expected MD5 hash in output, got: %s", stdout.String())
	}
}