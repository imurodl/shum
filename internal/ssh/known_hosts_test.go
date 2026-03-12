package ssh

import (
	"strings"
	"testing"
)

func TestExtractFingerprint(t *testing.T) {
	out := []byte("|1|example|example ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBDo4U9MMH9AQaJzYduH26WmQpzAv0MLoli0HApnJehs")
	fp := extractFingerprint(out)
	if !strings.HasPrefix(fp, "SHA256:") {
		t.Fatalf("expected a sha256 fingerprint, got %q", fp)
	}
}
