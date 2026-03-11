package ssh

import (
	"fmt"
	"strings"

	"github.com/your-org/shum/internal/remote"
)

type ProbeResult struct {
	OS            string
	Arch          string
	DockerVersion string
	ComposeVersion string
}

func ProbeAlias(alias string, runner *remote.Runner) (ProbeResult, error) {
	cmd := "uname -s; uname -m; command -v docker >/dev/null && docker --version && docker compose version"
	out, err := runner.Command(alias, cmd)
	if err != nil {
		return ProbeResult{}, err
	}
	lines := splitNonEmptyLines(out)
	if len(lines) < 4 {
		return ProbeResult{}, fmt.Errorf("probe output missing required lines")
	}
	return ProbeResult{
		OS:             lines[0],
		Arch:           lines[1],
		DockerVersion:  lines[2],
		ComposeVersion: lines[3],
	}, nil
}

func splitNonEmptyLines(raw string) []string {
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}
