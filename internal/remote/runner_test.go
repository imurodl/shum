package remote

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewRunnerDefaultTimeout(t *testing.T) {
	if got := NewRunner(0).Timeout; got != 20*time.Second {
		t.Errorf("default Timeout = %v, want 20s", got)
	}
	if got := NewRunner(5 * time.Second).Timeout; got != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", got)
	}
}

func TestCommandBuildsSSHInvocation(t *testing.T) {
	r := NewRunner(time.Second)
	cmd, cancel := r.command("prod", "docker compose ls")
	defer cancel()

	if base := filepath.Base(cmd.Path); base != "ssh" {
		t.Errorf("cmd.Path basename = %q, want ssh", base)
	}

	joined := strings.Join(cmd.Args, " ")
	for _, want := range []string{
		"BatchMode=yes",
		"StrictHostKeyChecking=yes",
		"prod",
		"docker compose ls",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("ssh args %v missing %q", cmd.Args, want)
		}
	}

	// The alias must come before the remote command in the argument list.
	aliasIdx, cmdIdx := indexOf(cmd.Args, "prod"), indexOf(cmd.Args, "docker compose ls")
	if aliasIdx == -1 || cmdIdx == -1 || aliasIdx > cmdIdx {
		t.Errorf("expected alias before command, got args %v", cmd.Args)
	}
}

func indexOf(args []string, target string) int {
	for i, a := range args {
		if a == target {
			return i
		}
	}
	return -1
}
