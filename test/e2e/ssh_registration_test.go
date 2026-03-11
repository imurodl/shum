package e2e

import (
	"os"
	"testing"
)

func TestSSHRegistrationRemoteGate(t *testing.T) {
	if alias := os.Getenv("SHUM_E2E_SSH_ALIAS"); alias == "" {
		t.Skip("SHUM_E2E_SSH_ALIAS is required for remote SSH E2E")
	}
}
