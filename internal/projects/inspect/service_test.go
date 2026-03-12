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

func TestResolveProfileOutputs(t *testing.T) {
	declared, active := resolveProfileOutputs("green\nworkers\n", []string{"green"})
	if len(declared) != 2 || declared[0] != "green" || declared[1] != "workers" {
		t.Fatalf("unexpected declared profiles: %#v", declared)
	}
	if len(active) != 1 || active[0] != "green" {
		t.Fatalf("unexpected active profiles: %#v", active)
	}
}
