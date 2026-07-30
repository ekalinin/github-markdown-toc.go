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
	if cfg.GHToken != "" {
		t.Errorf("got token %q, want empty", cfg.GHToken)
	}
	if cfg.GHUrl != defaultGitHubURL {
		t.Errorf("got GitHub URL %q, want %q", cfg.GHUrl, defaultGitHubURL)
	}
	if cfg.GHVersion != version.GH_2024_03 {
		t.Errorf("got regexp version %q, want %q", cfg.GHVersion, version.GH_2024_03)
	}
}

func TestParseConfigReadsEnvironment(t *testing.T) {
	t.Setenv("GH_TOC_TOKEN", "environment-token")
	t.Setenv("GH_TOC_URL", "https://github.example/api/v3")

	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GHToken != "environment-token" {
		t.Errorf("got token %q, want environment token", cfg.GHToken)
	}
	if cfg.GHUrl != "https://github.example/api/v3" {
		t.Errorf("got GitHub URL %q, want environment URL", cfg.GHUrl)
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
		"one.md",
		"two.md",
	}
	cfg, err := parseConfig(args)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GHToken != "flag-token" {
		t.Errorf("got token %q, want flag token", cfg.GHToken)
	}
	if cfg.GHUrl != "https://flag.example/api/v3" {
		t.Errorf("got GitHub URL %q, want flag URL", cfg.GHUrl)
	}
	if cfg.GHVersion != version.GH_V0 {
		t.Errorf("got regexp version %q, want %q", cfg.GHVersion, version.GH_V0)
	}
	if !cfg.Serial {
		t.Error("serial mode is disabled")
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

	if cfg.GHToken != "" {
		t.Errorf("got token %q, want empty flag value", cfg.GHToken)
	}
	if cfg.GHUrl != "" {
		t.Errorf("got GitHub URL %q, want empty flag value", cfg.GHUrl)
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
			if cfg.GHVersion != supported {
				t.Errorf("got regexp version %q, want %q", cfg.GHVersion, supported)
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
