package utils

import (
	"strings"
)

// RemoveStuff trims spaces, removes new lines and code tag from a string.
func RemoveStuff(s string) string {
	res := strings.ReplaceAll(s, "\n", "")
	res = strings.ReplaceAll(res, "<code>", "")
	res = strings.ReplaceAll(res, "</code>", "")
	res = strings.TrimSpace(res)

	return res
}
