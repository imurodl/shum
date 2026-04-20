// Package shumerr defines the stable error contract that shum exposes to
// callers — humans and AI agents alike. Error codes are part of the public
// surface: renaming one is a breaking change.
package shumerr

import (
	"errors"
	"fmt"
	"strings"
)

const (
	CodeInternal           = "internal_error"
	CodeUsage              = "usage_error"
	CodeHostUnreachable    = "host_unreachable"
	CodeHostUnverified     = "host_unverified"
	CodeHostNotFound       = "host_not_found"
	CodeHostNotLinux       = "host_not_linux"
	CodeSSHConfigInvalid   = "ssh_config_invalid"
	CodeKnownHostsMissing  = "known_hosts_missing"
	CodeProjectNotFound    = "project_not_found"
	CodeComposeUnavailable = "compose_unavailable"
	CodePolicyMissing      = "policy_missing"
	CodeBackupRequired     = "backup_required"
	CodeBackupFailed       = "backup_failed"
	CodeRestoreFailed      = "restore_failed"
	CodeArtifactNotFound   = "artifact_not_found"
	CodeMigrationWarning   = "migration_warning"
	CodePreflightBlocked   = "preflight_blocked"
	CodeUpgradeFailed      = "upgrade_failed"
	CodeRollbackFailed     = "rollback_failed"
	CodeHealthCheckFailed  = "health_check_failed"
	CodeProbeInvalid       = "probe_invalid"
	CodeStoreFailure       = "store_failure"
)

// Description returns a one-line description of an error code, used by
// `shum --agent-help` so agents can load all known codes into context once.
func Description(code string) string {
	switch code {
	case CodeInternal:
		return "Unclassified failure inside shum. File a bug if reproducible."
	case CodeUsage:
		return "Caller invoked shum with invalid or missing arguments."
	case CodeHostUnreachable:
		return "SSH could not reach the host (network, auth, or sshd down)."
	case CodeHostUnverified:
		return "Host key fingerprint did not match a trusted known_hosts entry."
	case CodeHostNotFound:
		return "No host registered under the given alias."
	case CodeHostNotLinux:
		return "Target host is not running Linux; shum only supports Linux hosts."
	case CodeSSHConfigInvalid:
		return "SSH config could not resolve the alias to a host."
	case CodeKnownHostsMissing:
		return "No known_hosts file is configured for the SSH alias."
	case CodeProjectNotFound:
		return "No compose project registered with the given reference on this host."
	case CodeComposeUnavailable:
		return "Docker or docker compose is not installed or not on PATH on the remote host."
	case CodePolicyMissing:
		return "Project has no safety policy configured. Run `shum project policy set` first."
	case CodeBackupRequired:
		return "Project policy requires a backup command but none is configured."
	case CodeBackupFailed:
		return "Backup command exited non-zero on the remote host."
	case CodeRestoreFailed:
		return "Restore command exited non-zero during rollback."
	case CodeArtifactNotFound:
		return "Backup artifact path does not exist."
	case CodeMigrationWarning:
		return "Project policy flags this upgrade as risky. Re-run with --force to proceed."
	case CodePreflightBlocked:
		return "Preflight checks reported one or more blocking issues."
	case CodeUpgradeFailed:
		return "Compose pull/up step failed on the remote host."
	case CodeRollbackFailed:
		return "Rollback after a failed upgrade also failed; manual intervention required."
	case CodeHealthCheckFailed:
		return "Post-upgrade health probes did not pass within the timeout."
	case CodeProbeInvalid:
		return "Health probe specification could not be parsed."
	case CodeStoreFailure:
		return "Local SQLite store returned an error."
	default:
		return ""
	}
}

// AllCodes returns every error code shum can emit, in stable order.
// Used by `shum --agent-help`.
func AllCodes() []string {
	return []string{
		CodeInternal,
		CodeUsage,
		CodeHostUnreachable,
		CodeHostUnverified,
		CodeHostNotFound,
		CodeHostNotLinux,
		CodeSSHConfigInvalid,
		CodeKnownHostsMissing,
		CodeProjectNotFound,
		CodeComposeUnavailable,
		CodePolicyMissing,
		CodeBackupRequired,
		CodeBackupFailed,
		CodeRestoreFailed,
		CodeArtifactNotFound,
		CodeMigrationWarning,
		CodePreflightBlocked,
		CodeUpgradeFailed,
		CodeRollbackFailed,
		CodeHealthCheckFailed,
		CodeProbeInvalid,
		CodeStoreFailure,
	}
}

// Error is the typed error shum returns from RunE handlers and from
// service-layer code that wants to make a code commitment.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Hint    string         `json:"hint,omitempty"`
	Details map[string]any `json:"details,omitempty"`
	cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.cause
}

// New returns a new typed error with the given code and message.
func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Newf is New with fmt formatting.
func Newf(code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap returns a typed error that wraps cause. If message is empty,
// the cause's message is used.
func Wrap(code string, cause error, message string) *Error {
	if cause == nil {
		return nil
	}
	if message == "" {
		message = cause.Error()
	}
	return &Error{Code: code, Message: message, cause: cause}
}

// WithHint attaches a remediation hint and returns e for chaining.
func (e *Error) WithHint(hint string) *Error {
	e.Hint = hint
	return e
}

// WithDetails attaches structured details and returns e for chaining.
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// From extracts a *Error from anywhere in the error chain.
func From(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var se *Error
	if errors.As(err, &se) {
		return se, true
	}
	return nil, false
}

// Classify converts an arbitrary error into a typed *Error. If the error
// already carries a code, that code is preserved; otherwise the message is
// inspected for known substrings and a best-effort code is assigned.
//
// New error sites should call New/Wrap directly; Classify exists so the
// boundary in cmd/shum/main.go works while older code is migrated.
func Classify(err error) *Error {
	if err == nil {
		return nil
	}
	if se, ok := From(err); ok {
		return se
	}
	msg := err.Error()
	code := CodeInternal
	switch {
	case strings.Contains(msg, "ssh connection to"):
		code = CodeHostUnreachable
	case strings.Contains(msg, "no known_hosts file"):
		code = CodeKnownHostsMissing
	case strings.Contains(msg, "alias resolution failed"),
		strings.Contains(msg, "ssh config"):
		code = CodeSSHConfigInvalid
	case strings.Contains(msg, "host probe failed"):
		code = CodeHostUnreachable
	case strings.Contains(msg, "target is not Linux"):
		code = CodeHostNotLinux
	case strings.Contains(msg, "no backup command configured"),
		strings.Contains(msg, "no restore command configured"):
		code = CodeBackupRequired
	case strings.Contains(msg, "backup command failed"):
		code = CodeBackupFailed
	case strings.Contains(msg, "restore command failed"):
		code = CodeRestoreFailed
	case strings.Contains(msg, "artifact not found"):
		code = CodeArtifactNotFound
	case strings.Contains(msg, "migration warning is enabled"):
		code = CodeMigrationWarning
	case strings.Contains(msg, "compose pull failed"),
		strings.Contains(msg, "compose up failed"):
		code = CodeUpgradeFailed
	case strings.Contains(msg, "health verification failed"),
		strings.Contains(msg, "http probe failed"),
		strings.Contains(msg, "tcp probe failed"),
		strings.Contains(msg, "cmd probe failed"):
		code = CodeHealthCheckFailed
	case strings.Contains(msg, "host alias and project ref required"):
		code = CodeUsage
	case strings.Contains(msg, "invalid health check"),
		strings.Contains(msg, "missing target in health check"),
		strings.Contains(msg, "unsupported health check type"):
		code = CodeProbeInvalid
	}
	return &Error{Code: code, Message: msg, cause: err}
}

// ExitCode returns the canonical process exit code for an error code.
// Documented as part of the agent contract.
func ExitCode(code string) int {
	switch code {
	case CodeUsage, CodeProbeInvalid:
		return 64
	case CodeHostUnreachable, CodeHostUnverified, CodeHostNotFound, CodeHostNotLinux,
		CodeSSHConfigInvalid, CodeKnownHostsMissing:
		return 65
	case CodePreflightBlocked, CodeComposeUnavailable, CodeProjectNotFound, CodePolicyMissing:
		return 66
	case CodeBackupRequired, CodeBackupFailed, CodeRestoreFailed, CodeArtifactNotFound:
		return 67
	case CodeUpgradeFailed, CodeMigrationWarning, CodeHealthCheckFailed, CodeRollbackFailed:
		return 68
	case CodeStoreFailure:
		return 70
	default:
		return 1
	}
}
