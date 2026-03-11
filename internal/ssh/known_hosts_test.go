package ssh

import "testing"

func TestExtractFingerprint(t *testing.T) {
	out := []byte("127.0.0.1 ssh-rsa AAAAB3NzaC1yc2E AAAAB3NzaC1yc2E comment")
	fp := extractFingerprint(out)
	if fp == "" {
		t.Fatalf("expected a parsed fingerprint")
	}
}
