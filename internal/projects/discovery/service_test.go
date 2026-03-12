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

func TestParseComposeLSOutputArray(t *testing.T) {
	raw := `[{"Name":"politech","ConfigFiles":"/home/ubuntu/politech/docker-compose.yml"},{"Name":"rss-proxy","ConfigFiles":"/home/ubuntu/rss-proxy/docker-compose.yml"}]`

	got := parseComposeLSOutput(raw)
	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	if got[0].Name != "politech" {
		t.Fatalf("unexpected first project: %+v", got[0])
	}
	if got[0].Directory != "/home/ubuntu/politech" {
		t.Fatalf("unexpected project directory: %+v", got[0])
	}
	if len(got[0].ComposeFiles) != 1 || got[0].ComposeFiles[0] != "/home/ubuntu/politech/docker-compose.yml" {
		t.Fatalf("unexpected compose files: %+v", got[0].ComposeFiles)
	}
}

func TestParseComposeLSOutputLineDelimited(t *testing.T) {
	raw := "{\"Name\":\"politech\",\"ConfigFiles\":\"/home/ubuntu/politech/docker-compose.yml\"}\n{\"Name\":\"rss-proxy\",\"ConfigFiles\":\"/home/ubuntu/rss-proxy/docker-compose.yml\"}\n"

	got := parseComposeLSOutput(raw)
	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	if got[1].Name != "rss-proxy" {
		t.Fatalf("unexpected second project: %+v", got[1])
	}
}
