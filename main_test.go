package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunStdinHappyPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("│ hello\n│ world\n")
	code := run([]string{"--stdin"}, in, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code: got %d want 0; stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "hello\nworld\n" {
		t.Errorf("stdout: got %q want %q", got, "hello\nworld\n")
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr: got %q want empty", got)
	}
}

func TestRunStdinDryRunConflict(t *testing.T) {
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("anything")
	code := run([]string{"--stdin", "--dry-run"}, in, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code: got %d want 2", code)
	}
	if !strings.Contains(stderr.String(), "--dry-run applies to clipboard mode only") {
		t.Errorf("stderr should explain conflict; got %q", stderr.String())
	}
}

func TestRunStdinExplain(t *testing.T) {
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("│ a\n│ b\n│ c")
	code := run([]string{"--stdin", "--explain"}, in, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code: got %d want 0", code)
	}
	if got := stdout.String(); got != "a\nb\nc" {
		t.Errorf("stdout: got %q want %q", got, "a\nb\nc")
	}
	se := stderr.String()
	if !strings.Contains(se, "ai-clean:") {
		t.Errorf("stderr missing summary header: %q", se)
	}
	if !strings.Contains(se, "leading border") {
		t.Errorf("stderr missing leading-border summary: %q", se)
	}
	if !strings.Contains(se, "stripped from 3 line(s)") {
		t.Errorf("stderr missing line count: %q", se)
	}
}

func TestRunStdinExplainNoChanges(t *testing.T) {
	var stdout, stderr bytes.Buffer
	in := strings.NewReader("plain text\n")
	code := run([]string{"--stdin", "--explain"}, in, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code: got %d want 0", code)
	}
	if !strings.Contains(stderr.String(), "no changes") {
		t.Errorf("stderr should report no-changes; got %q", stderr.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code: got %d want 0", code)
	}
	if got := strings.TrimSpace(stdout.String()); got == "" {
		t.Errorf("stdout should print version; got empty")
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--bogus"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Errorf("unknown flag should exit 2; got %d", code)
	}
}
