package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestFullStartsWithBareVersion(t *testing.T) {
	lines := strings.Split(Full(), "\n")
	if lines[0] != Version {
		t.Errorf("got first line %q, want the bare version %q", lines[0], Version)
	}
}

func TestFullReportsPlatform(t *testing.T) {
	full := Full()
	for _, want := range []string{runtime.GOOS, runtime.GOARCH, runtime.Version()} {
		if !strings.Contains(full, want) {
			t.Errorf("got %q, want it to contain %q", full, want)
		}
	}
}
