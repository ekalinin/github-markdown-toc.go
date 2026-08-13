package adapters

import (
	"os"
	"path/filepath"
	"strings"
)

const tokenFileName = "token.txt"

// TokenResolver reads the GitHub token from a file next to the executable. It is the
// last fallback, after the --token flag and the GH_TOC_TOKEN environment variable.
type TokenResolver struct {
	executable func() (string, error)
}

func NewTokenResolver() *TokenResolver {
	return NewTokenResolverX(os.Executable)
}

func NewTokenResolverX(executable func() (string, error)) *TokenResolver {
	return &TokenResolver{executable: executable}
}

// Resolve returns the token found in token.txt, or an empty string when there is no
// such file. A missing file is not an error.
func (r *TokenResolver) Resolve() (string, error) {
	path, err := r.executable()
	if err != nil {
		return "", nil
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(path), tokenFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
