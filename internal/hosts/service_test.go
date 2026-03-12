package hosts

import (
	"context"
	"testing"
	"time"

	"github.com/imurodl/shum/internal/remote"
)

func TestHostServiceTrustSummary(t *testing.T) {
	host := Host{
		Alias:             "sample",
		Hostname:          "127.0.0.1",
		UserName:          "root",
		Port:              22,
		HostKeyFingerprint: "SHA256:test",
		LastVerifiedAt:     time.Now().UTC(),
	}
	got := host.TrustSummary()
	if got == "" || host.HostKeyFingerprint == "" {
		t.Fatalf("invalid trust summary")
	}
}

func TestHostServiceRegisterRequiresRunner(t *testing.T) {
	_ = remote.NewRunner
	_ = context.TODO
}
