package discovery

import "testing"

func TestStatusFromComposeState(t *testing.T) {
	got := StatusFromComposeState(0)
	if got != "canonical" {
		t.Fatalf("expected canonical, got %s", got)
	}
}

func TestRenderCountByStatus(t *testing.T) {
	counts := RenderCountByStatus([]RuntimeProject{
		{Name: "a", Status: "runtime_only"},
		{Name: "b", Status: "runtime_only"},
	})
	if counts["runtime_only"] != 2 {
		t.Fatalf("unexpected counts: %v", counts)
	}
}
