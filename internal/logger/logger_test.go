package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewTextFormat(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "debug", Format: "text", Output: &buf})
	log.Debug("hello", "key", "value")

	out := buf.String()
	if !strings.Contains(out, "hello") || !strings.Contains(out, "key=value") {
		t.Fatalf("unexpected text log output: %q", out)
	}
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "info", Format: "json", Output: &buf})
	log.Info("hello", "key", "value")

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON log line: %v", err)
	}
	if payload["msg"] != "hello" || payload["key"] != "value" {
		t.Fatalf("unexpected JSON payload: %+v", payload)
	}
}

func TestNewFiltersBelowLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(Config{Level: "error", Format: "text", Output: &buf})
	log.Debug("hidden")

	if buf.Len() != 0 {
		t.Fatalf("expected no output below level, got %q", buf.String())
	}
}
