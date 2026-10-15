package problemid

import (
	"fmt"
	"regexp"
	"strings"
)

var format = regexp.MustCompile(`^(\d+)([A-Z]\d*)$`)

func Parse(raw string) (string, string, error) {
	matches := format.FindStringSubmatch(strings.ToUpper(raw))
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid problem ID format '%s'. use format like '71A' or '4A'", raw)
	}
	return matches[1], matches[2], nil
}
