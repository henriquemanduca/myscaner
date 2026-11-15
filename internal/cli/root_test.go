package cli

import (
	"bytes"
	"strings"
	"testing"
)

func isolateHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
}

func TestRootCommandVersionFlag(t *testing.T) {
	isolateHome(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("executing --version: %v", err)
	}

	if got := strings.TrimSpace(out.String()); !strings.HasPrefix(got, "scaner-cli ") {
		t.Fatalf("unexpected version output: %q", got)
	}
}

func TestRootCommandWithoutInputPrintsHelp(t *testing.T) {
	isolateHome(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("executing without args: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}
