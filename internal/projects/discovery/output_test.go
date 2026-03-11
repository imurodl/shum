package discovery

import "testing"

func TestRenderDiscoverJSON(t *testing.T) {
	raw := RenderDiscoverJSON([]RuntimeProject{
		{Name: "alpha", Status: "runtime_only", Source: "compose ls"},
	})
	if raw == "" {
		t.Fatalf("expected summary output")
	}
}
