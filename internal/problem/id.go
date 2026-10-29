package problem

import (
	"fmt"
	"regexp"
	"strings"
)

var idFormat = regexp.MustCompile(`^(\d+)([A-Z]\d*)$`)

func ParseID(raw string) (string, string, error) {
	matches := idFormat.FindStringSubmatch(strings.ToUpper(raw))
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid problem ID format '%s'. use format like '71A' or '4A'", raw)
	}
	return matches[1], matches[2], nil
}
