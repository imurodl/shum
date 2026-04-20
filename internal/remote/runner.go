package remote

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/imurodl/shum/internal/shumerr"
)

type Runner struct {
	Timeout time.Duration
}

func NewRunner(timeout time.Duration) *Runner {
	if timeout == 0 {
		timeout = 20 * time.Second
	}
	return &Runner{Timeout: timeout}
}

func (r *Runner) Command(alias string, command string) (string, error) {
	cmd, cancel := r.command(alias, command)
	defer cancel()
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 255 {
			return "", shumerr.Newf(shumerr.CodeHostUnreachable, "ssh connection to %q failed: %s", alias, output).
				WithHint("verify the host is up, the SSH alias resolves, and the key is trusted").
				WithDetails(map[string]any{"alias": alias, "ssh_output": output})
		}
		return "", shumerr.Wrap(shumerr.CodeInternal, err, fmt.Sprintf("remote command on %q failed: %s", alias, output)).
			WithDetails(map[string]any{"alias": alias, "ssh_output": output})
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *Runner) command(alias string, command string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	cmd := exec.CommandContext(ctx, "ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=yes",
		alias,
		command,
	)
	return cmd, cancel
}
