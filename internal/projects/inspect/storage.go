package inspect

import (
	"strings"
)

func ParseServiceStates(raw string) []string {
	out := []string{}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}
