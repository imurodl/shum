package inspect

import (
	"testing"
)

func TestEnvFingerprintDeterministic(t *testing.T) {
	a := envFingerprint("a\nb")
	b := envFingerprint("a\nb")
	if a != b {
		t.Fatalf("expected deterministic fingerprint")
	}
}
