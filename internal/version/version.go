package version

import (
	"fmt"
	"runtime"
)

const (
	// Version is a current app version
	Version   = "2.0.1"
	UserAgent = "github-markdown-toc.go v" + Version
)

// Versions of GH layouts
const (
	GH_V0      = "0"
	GH_2023_10 = "2023-10"
	GH_2024_03 = "2024-03"
)

// SupportedGHVersions returns the supported GitHub layout versions.
func SupportedGHVersions() []string {
	return []string{
		GH_V0,
		GH_2023_10,
		GH_2024_03,
	}
}

// Full returns the multi-line version banner. The first line stays the bare version
// number, so scripts that parse `gh-md-toc --version` keep working.
func Full() string {
	return fmt.Sprintf("%s\n\nos:   %s\narch: %s\ngo:   %s",
		Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
