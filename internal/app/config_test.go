package app

import (
	"slices"
	"testing"
)

func Test_ConfigToControllerConfig(t *testing.T) {
	cfg := Config{
		Files:      []string{"f1", "f2"},
		Serial:     true,
		HideHeader: true,
		HideFooter: true,
		StartDepth: 10,
		Depth:      20,
		NoEscape:   true,
		Indent:     15,
		Debug:      true,
		GHToken:    "t1",
		GHUrl:      "some-url",
		GHVersion:  "some-version",
	}
	cfgCtrl := cfg.ToControllerConfig()

	if !slices.Equal(cfg.Files, cfgCtrl.Files) {
		t.Errorf("Files are not the same. Got=%v, want=%v\n", cfgCtrl.Files, cfg.Files)
	}

	if cfg.Serial != cfgCtrl.Serial {
		t.Errorf("Serial is not the same. Got=%v, want=%v\n", cfgCtrl.Serial, cfg.Serial)
	}

	if cfg.HideHeader != cfgCtrl.HideHeader {
		t.Errorf("HideHeader is not the same. Got=%v, want=%v\n", cfgCtrl.HideHeader, cfg.HideHeader)
	}

	if cfg.HideFooter != cfgCtrl.HideFooter {
		t.Errorf("HideFooter is not the same. Got=%v, want=%v\n", cfgCtrl.HideFooter, cfg.HideFooter)
	}

	if cfg.StartDepth != cfgCtrl.StartDepth {
		t.Errorf("StartDepth is not the same. Got=%v, want=%v\n", cfgCtrl.StartDepth, cfg.StartDepth)
	}

	if cfg.Depth != cfgCtrl.Depth {
		t.Errorf("Depth is not the same. Got=%v, want=%v\n", cfgCtrl.Depth, cfg.Depth)
	}

	if cfg.NoEscape != cfgCtrl.NoEscape {
		t.Errorf("NoEscape is not the same. Got=%v, want=%v\n", cfgCtrl.NoEscape, cfg.NoEscape)
	}

	if cfg.Indent != cfgCtrl.Indent {
		t.Errorf("Indent is not the same. Got=%v, want=%v\n", cfgCtrl.Indent, cfg.Indent)
	}

	if cfg.Debug != cfgCtrl.Debug {
		t.Errorf("Debug is not the same. Got=%v, want=%v\n", cfgCtrl.Debug, cfg.Debug)
	}

	if cfg.GHToken != cfgCtrl.GHToken {
		t.Errorf("GHToken is not the same. Got=%v, want=%v\n", cfgCtrl.GHToken, cfg.GHToken)
	}

	if cfg.GHUrl != cfgCtrl.GHUrl {
		t.Errorf("GHUrl is not the same. Got=%v, want=%v\n", cfgCtrl.GHUrl, cfg.GHUrl)
	}

	if cfg.GHVersion != cfgCtrl.GHVersion {
		t.Errorf("GHVersion is not the same. Got=%v, want=%v\n", cfgCtrl.GHVersion, cfg.GHVersion)
	}
}

func Test_ConfigToGrabberConfig(t *testing.T) {
	cfg := Config{
		Files:      []string{"f1", "f2"},
		Serial:     true,
		HideHeader: true,
		HideFooter: true,
		StartDepth: 10,
		Depth:      20,
		NoEscape:   true,
		Indent:     15,
		Debug:      true,
		GHToken:    "t1",
		GHUrl:      "some-url",
		GHVersion:  "some-version",
	}
	cfgGrbr := cfg.ToGrabberConfig()

	if cfg.StartDepth != cfgGrbr.StartDepth {
		t.Errorf("StartDepth is not the same. Got=%v, want=%v\n", cfgGrbr.StartDepth, cfg.StartDepth)
	}

	if cfg.Depth != cfgGrbr.Depth {
		t.Errorf("Depth is not the same. Got=%v, want=%v\n", cfgGrbr.Depth, cfg.Depth)
	}

	if !cfgGrbr.AbsPaths {
		t.Errorf("AbsPaths should be true. Got=%v\n", cfgGrbr.AbsPaths)
	}

	if cfg.NoEscape == cfgGrbr.Escape {
		t.Errorf("NoEscape is the same. Got=%v, want=%v\n", cfgGrbr.Escape, cfg.NoEscape)
	}

	if cfg.Indent != cfgGrbr.Indent {
		t.Errorf("Indent is not the same. Got=%v, want=%v\n", cfgGrbr.Indent, cfg.Indent)
	}
}
