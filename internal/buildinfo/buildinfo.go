package buildinfo

import (
	"fmt"
	"runtime"
)

var (
	// Version holds the current semantic version. Injected via ldflags during build.
	Version = "dev"
	// Commit holds the git commit hash. Injected via ldflags during build.
	Commit = "none"
	// Date holds the build timestamp. Injected via ldflags during build.
	Date = "unknown"
)

// Info represents full build metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns current build information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String formats the build information as a human-readable string.
func (i Info) String() string {
	return fmt.Sprintf("scaner-cli %s (commit: %s, built at: %s, %s/%s, %s)",
		i.Version, i.Commit, i.Date, i.OS, i.Arch, i.GoVersion)
}
