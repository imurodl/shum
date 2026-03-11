package ops

import "testing"

func TestExtractServiceChangesArrayJSON(t *testing.T) {
	raw := `[{"Service":"web","Image":"nginx:1.27","State":"running","Health":"healthy"},{"Service":"api","Image":"app:latest","State":"exited","Health":"unhealthy"}]`
	services := extractServiceChanges(raw)
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
	if services[0].ServiceName != "web" {
		t.Fatalf("expected web service, got %q", services[0].ServiceName)
	}
	if services[1].CurrentHealthy != "exited" {
		t.Fatalf("expected exited health marker, got %q", services[1].CurrentHealthy)
	}
}

func TestExtractServiceChangesLineJSON(t *testing.T) {
	raw := "{\"Name\":\"web\",\"Image\":\"nginx:1.27\",\"State\":\"running\"}\n{\"Name\":\"api\",\"Image\":\"worker:1.0\",\"State\":\"up\"}\n"
	services := extractServiceChanges(raw)
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
	if services[1].ServiceName != "api" {
		t.Fatalf("expected api service, got %q", services[1].ServiceName)
	}
}

func TestExtractServiceChangesMapFallback(t *testing.T) {
	raw := `{"web":{"Service":"web","Image":"nginx","State":"running","Health":"healthy"}}`
	var rows []composePSLine
	if err := parseComposePS(raw, &rows); err != nil {
		t.Fatalf("failed to parse map fallback: %v", err)
	}
	t.Logf("rows=%#v", rows)
	services := extractServiceChanges(raw)
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d for raw %s", len(services), raw)
	}
	if services[0].ServiceName != "web" {
		t.Fatalf("expected web service, got %q", services[0].ServiceName)
	}
}
