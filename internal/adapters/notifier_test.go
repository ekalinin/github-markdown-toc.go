package adapters

import (
	"bytes"
	"testing"
)

func TestNotifierNotify(t *testing.T) {
	var buf bytes.Buffer
	NewNotifier(&buf).Notify("!! TOC was added into: %q", "README.md")

	want := "!! TOC was added into: \"README.md\"\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestNotifierNilWriter(t *testing.T) {
	NewNotifier(nil).Notify("ignored %s", "message")
}
