package adapters

import (
	"errors"
	"testing"
	"time"
)

func TestStamperStamp(t *testing.T) {
	stamp := time.Date(2026, 8, 12, 13, 45, 6, 0, time.UTC)
	tests := []struct {
		name     string
		username func() (string, error)
		want     string
	}{
		{
			name:     "known user",
			username: func() (string, error) { return "ekalinin", nil },
			want:     "<!-- Added by: ekalinin, at: 2026-08-12T13:45:06Z -->",
		},
		{
			name:     "lookup fails",
			username: func() (string, error) { return "", errors.New("no user") },
			want:     "<!-- Added by: unknown, at: 2026-08-12T13:45:06Z -->",
		},
		{
			name:     "empty user",
			username: func() (string, error) { return "", nil },
			want:     "<!-- Added by: unknown, at: 2026-08-12T13:45:06Z -->",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStamperX(func() time.Time { return stamp }, tt.username).Stamp()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
