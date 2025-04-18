package entity

import (
	"net/url"
	"strings"
)

type Type int

const (
	TypeLocalMD Type = iota
	TypeRemoteMD
	TypeRemoteHTML
)

func GetType(path string) Type {
	u, err := url.Parse(path)
	if err != nil || u.Scheme == "" {
		return TypeLocalMD
	}
	if strings.Contains(path, "githubusercontent.com") {
		return TypeRemoteMD
	}
	return TypeRemoteHTML
}
