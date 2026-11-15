package buildinfo

import (
	"strings"
	"testing"
)

func TestGetIncludesRuntimeMetadata(t *testing.T) {
	info := Get()
	if info.Version == "" || info.Commit == "" || info.Date == "" {
		t.Fatalf("build metadata must not be empty: %+v", info)
	}
	if info.GoVersion == "" || info.OS == "" || info.Arch == "" {
		t.Fatalf("runtime metadata must not be empty: %+v", info)
	}
}

func TestString(t *testing.T) {
	info := Info{Version: "v1.2.3", Commit: "abc123", Date: "2026-01-01T00:00:00Z"}
	s := info.String()

	for _, want := range []string{"scaner-cli v1.2.3", "abc123", "2026-01-01T00:00:00Z"} {
		if !strings.Contains(s, want) {
			t.Fatalf("String() = %q, missing %q", s, want)
		}
	}
}
