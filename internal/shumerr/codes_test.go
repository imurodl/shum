package shumerr

import (
	"errors"
	"testing"
)

func TestNewSetsCodeAndMessage(t *testing.T) {
	e := New(CodeHostUnreachable, "boom")
	if e.Code != CodeHostUnreachable {
		t.Fatalf("code = %s, want %s", e.Code, CodeHostUnreachable)
	}
	if e.Error() != "boom" {
		t.Fatalf("Error() = %q, want %q", e.Error(), "boom")
	}
}

func TestWrapPreservesCause(t *testing.T) {
	cause := errors.New("underneath")
	e := Wrap(CodeBackupFailed, cause, "outer")
	if !errors.Is(e, cause) {
		t.Fatalf("errors.Is should match wrapped cause")
	}
	if e.Code != CodeBackupFailed {
		t.Fatalf("code = %s, want %s", e.Code, CodeBackupFailed)
	}
}

func TestWrapDefaultsMessageToCause(t *testing.T) {
	cause := errors.New("from below")
	e := Wrap(CodeInternal, cause, "")
	if e.Message != "from below" {
		t.Fatalf("expected cause message to be reused, got %q", e.Message)
	}
}

func TestFromExtractsThroughErrorChain(t *testing.T) {
	original := New(CodeHostUnreachable, "ssh down")
	wrapped := errors.New("outer: " + original.Error())
	// errors.As only finds within the chain; this check is for direct usage
	if got, ok := From(original); !ok || got.Code != CodeHostUnreachable {
		t.Fatalf("From(original) failed to extract")
	}
	// And via fmt.Errorf style wrapping
	wrapped2 := wrapErr(original)
	if got, ok := From(wrapped2); !ok || got.Code != CodeHostUnreachable {
		t.Fatalf("From(wrapped) failed to extract through chain")
	}
	_ = wrapped
}

func wrapErr(err error) error {
	type holder struct{ inner error }
	// use a wrapping error type to ensure errors.As traversal
	return &chainedErr{inner: err}
}

type chainedErr struct{ inner error }

func (c *chainedErr) Error() string { return "chain: " + c.inner.Error() }
func (c *chainedErr) Unwrap() error { return c.inner }

func TestClassifyKnownStrings(t *testing.T) {
	tests := []struct {
		msg  string
		want string
	}{
		{"ssh connection to \"prod\" failed: network unreachable", CodeHostUnreachable},
		{"no known_hosts file configured for ssh alias prod", CodeKnownHostsMissing},
		{"alias resolution failed: not in config", CodeSSHConfigInvalid},
		{"target is not Linux: darwin", CodeHostNotLinux},
		{"no backup command configured", CodeBackupRequired},
		{"backup command failed: exit 2", CodeBackupFailed},
		{"restore command failed: bad", CodeRestoreFailed},
		{"artifact not found: /tmp/x", CodeArtifactNotFound},
		{"migration warning is enabled; use --force to continue", CodeMigrationWarning},
		{"compose pull failed: timeout", CodeUpgradeFailed},
		{"health verification failed: probe down", CodeHealthCheckFailed},
		{"invalid health check format: bad", CodeProbeInvalid},
		{"some random thing", CodeInternal},
	}
	for _, tc := range tests {
		got := Classify(errors.New(tc.msg))
		if got.Code != tc.want {
			t.Errorf("Classify(%q) code = %s, want %s", tc.msg, got.Code, tc.want)
		}
	}
}

func TestClassifyPreservesTypedError(t *testing.T) {
	e := New(CodePolicyMissing, "no policy")
	got := Classify(e)
	if got.Code != CodePolicyMissing {
		t.Fatalf("Classify must preserve typed code, got %s", got.Code)
	}
}

func TestExitCodesAreStable(t *testing.T) {
	// These exit codes are part of the agent contract. Changing them is breaking.
	cases := map[string]int{
		CodeUsage:             64,
		CodeProbeInvalid:      64,
		CodeHostUnreachable:   65,
		CodeHostNotFound:      65,
		CodeKnownHostsMissing: 65,
		CodePreflightBlocked:  66,
		CodeProjectNotFound:   66,
		CodeBackupRequired:    67,
		CodeBackupFailed:      67,
		CodeArtifactNotFound:  67,
		CodeUpgradeFailed:     68,
		CodeMigrationWarning:  68,
		CodeRollbackFailed:    68,
		CodeHealthCheckFailed: 68,
		CodeStoreFailure:      70,
		CodeInternal:          1,
	}
	for code, want := range cases {
		if got := ExitCode(code); got != want {
			t.Errorf("ExitCode(%s) = %d, want %d (CONTRACT CHANGE: bump revision)", code, got, want)
		}
	}
}

func TestAllCodesHaveDescriptions(t *testing.T) {
	for _, code := range AllCodes() {
		if Description(code) == "" {
			t.Errorf("code %q has no Description; agents need it for context", code)
		}
	}
}

func TestErrorChainable(t *testing.T) {
	e := New(CodeBackupFailed, "x").
		WithHint("try y").
		WithDetails(map[string]any{"k": "v"})
	if e.Hint != "try y" {
		t.Fatalf("hint not chained")
	}
	if e.Details["k"] != "v" {
		t.Fatalf("details not chained")
	}
}
