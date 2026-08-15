package adapters

import (
	"fmt"
	"os/user"
	"time"
)

// Stamper builds the signature comment written into a document next to an
// inserted TOC.
type Stamper struct {
	now      func() time.Time
	username func() (string, error)
}

func NewStamper() *Stamper {
	return NewStamperX(time.Now, currentUsername)
}

func NewStamperX(now func() time.Time, username func() (string, error)) *Stamper {
	return &Stamper{now: now, username: username}
}

func currentUsername() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", err
	}
	return current.Username, nil
}

func (s *Stamper) Stamp() string {
	name, err := s.username()
	if err != nil || name == "" {
		name = "unknown"
	}
	return fmt.Sprintf("<!-- Added by: %s, at: %s -->", name, s.now().Format(time.RFC3339))
}
