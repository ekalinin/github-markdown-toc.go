package main

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

func TestParseConfigDefaults(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "")
	t.Setenv("GH_TOC_URL", "")

	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Files) != 0 {
		t.Errorf("got files %v, want none", cfg.Files)
	}
	if cfg.GitHub.GHToken != "" {
		t.Errorf("got token %q, want empty", cfg.GitHub.GHToken)
	}
	if cfg.GitHub.GHUrl != defaultGitHubURL {
		t.Errorf("got GitHub URL %q, want %q", cfg.GitHub.GHUrl, defaultGitHubURL)
	}
	if cfg.GitHub.GHVersion != version.GH_2024_03 {
		t.Errorf("got regexp version %q, want %q", cfg.GitHub.GHVersion, version.GH_2024_03)
	}
}

func TestParseConfigReadsEnvironment(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "environment-token")
	t.Setenv("GH_TOC_URL", "https://github.example/api/v3")

	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GitHub.GHToken != "environment-token" {
		t.Errorf("got token %q, want environment token", cfg.GitHub.GHToken)
	}
	if cfg.GitHub.GHUrl != "https://github.example/api/v3" {
		t.Errorf("got GitHub URL %q, want environment URL", cfg.GitHub.GHUrl)
	}
}

func TestParseConfigFlagsOverrideEnvironment(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "environment-token")
	t.Setenv("GH_TOC_URL", "https://environment.example/api/v3")

	args := []string{
		"--token=flag-token",
		"--github-url=https://flag.example/api/v3",
		"--re-version=" + version.GH_V0,
		"--serial",
		"--hide-header",
		"--hide-footer",
		"--start-depth=2",
		"--depth=3",
		"--no-escape",
		"--indent=4",
		"--debug",
		"one.md",
		"two.md",
	}
	cfg, err := parseConfig(args)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GitHub.GHToken != "flag-token" {
		t.Errorf("got token %q, want flag token", cfg.GitHub.GHToken)
	}
	if cfg.GitHub.GHUrl != "https://flag.example/api/v3" {
		t.Errorf("got GitHub URL %q, want flag URL", cfg.GitHub.GHUrl)
	}
	if cfg.GitHub.GHVersion != version.GH_V0 {
		t.Errorf("got regexp version %q, want %q", cfg.GitHub.GHVersion, version.GH_V0)
	}
	if !cfg.Serial {
		t.Error("serial mode is disabled")
	}
	if !cfg.Presentation.HideHeader {
		t.Error("TOC header is not hidden")
	}
	if !cfg.Presentation.HideFooter {
		t.Error("TOC footer is not hidden")
	}
	if cfg.TOC.StartDepth != 2 {
		t.Errorf("got start depth %d, want 2", cfg.TOC.StartDepth)
	}
	if cfg.TOC.Depth != 3 {
		t.Errorf("got depth %d, want 3", cfg.TOC.Depth)
	}
	if cfg.TOC.Escape {
		t.Error("TOC escaping is enabled")
	}
	if cfg.TOC.Indent != 4 {
		t.Errorf("got indentation %d, want 4", cfg.TOC.Indent)
	}
	if !cfg.Debug {
		t.Error("debug mode is disabled")
	}
	if want := []string{"one.md", "two.md"}; !slices.Equal(cfg.Files, want) {
		t.Errorf("got files %v, want %v", cfg.Files, want)
	}
}

func TestParseConfigEmptyFlagsOverrideEnvironment(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "environment-token")
	t.Setenv("GH_TOC_URL", "https://environment.example/api/v3")

	cfg, err := parseConfig([]string{"--token=", "--github-url="})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GitHub.GHToken != "" {
		t.Errorf("got token %q, want empty flag value", cfg.GitHub.GHToken)
	}
	if cfg.GitHub.GHUrl != "" {
		t.Errorf("got GitHub URL %q, want empty flag value", cfg.GitHub.GHUrl)
	}
}

func TestParseConfigAcceptsSupportedRegexpVersions(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "")
	t.Setenv("GH_TOC_URL", "")

	for _, supported := range version.SupportedGHVersions() {
		t.Run(supported, func(t *testing.T) {
			cfg, err := parseConfig([]string{"--re-version=" + supported})
			if err != nil {
				t.Fatal(err)
			}
			if cfg.GitHub.GHVersion != supported {
				t.Errorf("got regexp version %q, want %q", cfg.GitHub.GHVersion, supported)
			}
		})
	}
}

func TestParseConfigRejectsUnknownRegexpVersion(t *testing.T) {
	_, err := parseConfig([]string{"--re-version=unknown"})
	if err == nil {
		t.Fatal("expected an error for an unsupported regexp version")
	}

	for _, expected := range []string{"unknown", "0", "2023-10", "2024-03"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("error %q does not contain %q", err, expected)
		}
	}
}

func TestCLIHelpShowsCurrentDefaultsWithoutToken(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "secret-token")
	t.Setenv("GH_TOC_URL", "https://github.example/api/v3")

	parser, _ := newCLI()
	var output bytes.Buffer
	parser.UsageWriter(&output)
	parser.Usage(nil)

	help := output.String()
	for _, expected := range []string{
		"GitHub URL. Default: " + defaultGitHubURL,
		"RegExp version. Default: " + version.GH_2024_03,
	} {
		if !strings.Contains(help, expected) {
			t.Errorf("help does not contain %q:\n%s", expected, help)
		}
	}
	if strings.Contains(help, "secret-token") {
		t.Errorf("help contains GitHub token:\n%s", help)
	}
}

func TestParseConfigStdinMarker(t *testing.T) {
	cfg, err := parseConfig([]string{"-"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Files) != 0 {
		t.Errorf("got files %v, want none so STDIN is used", cfg.Files)
	}
}

func TestParseConfigStdinMarkerWithPaths(t *testing.T) {
	_, err := parseConfig([]string{"-", "README.md"})
	if err == nil {
		t.Fatal("got no error, want a usage error")
	}
	if !strings.Contains(err.Error(), "STDIN marker") {
		t.Errorf("got error %q, want it to mention the STDIN marker", err)
	}
}

func TestParseConfigInsertFlags(t *testing.T) {
	cfg, err := parseConfig([]string{"--insert", "--no-backup", "README.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Insert.Enabled || !cfg.Insert.NoBackup {
		t.Errorf("got insert config %+v, want both flags set", cfg.Insert)
	}
}

func TestParseConfigNoBackupRequiresInsert(t *testing.T) {
	_, err := parseConfig([]string{"--no-backup", "README.md"})
	if err == nil {
		t.Fatal("got no error, want a usage error")
	}
	if !strings.Contains(err.Error(), "--no-backup requires --insert") {
		t.Errorf("got error %q, want it to explain the dependency", err)
	}
}

func TestParseConfigInsertRequiresFilePath(t *testing.T) {
	_, err := parseConfig([]string{"--insert"})
	if err == nil {
		t.Fatal("got no error, want a usage error")
	}
	if !strings.Contains(err.Error(), "--insert requires at least one file path") {
		t.Errorf("got error %q, want it to explain the dependency", err)
	}
}

func TestParseConfigInsertRejectsStdinMarker(t *testing.T) {
	_, err := parseConfig([]string{"--insert", "-"})
	if err == nil {
		t.Fatal("got no error, want a usage error")
	}
	if !strings.Contains(err.Error(), "--insert requires at least one file path") {
		t.Errorf("got error %q, want it to explain the dependency", err)
	}
}
