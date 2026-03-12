package cli

import (
	"testing"

	"github.com/imurodl/shum/internal/ops"
)

func TestParseHealthProbes(t *testing.T) {
	raw := []string{
		"http://localhost:8080/health",
		"tcp:127.0.0.1:5432",
		"cmd:docker ps",
	}
	probes, err := parseHealthProbes(raw)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if len(probes) != 3 {
		t.Fatalf("expected 3 probes, got %d", len(probes))
	}
	if probes[0].Type != "http" {
		t.Fatalf("expected first probe to be http, got %s", probes[0].Type)
	}
	if probes[1].Type != "tcp" {
		t.Fatalf("expected second probe to be tcp, got %s", probes[1].Type)
	}
	if probes[2].Type != "cmd" {
		t.Fatalf("expected third probe to be cmd, got %s", probes[2].Type)
	}
}

func TestParseHealthProbesRejectsBadInput(t *testing.T) {
	_, err := parseHealthProbes([]string{"badprobe"})
	if err == nil {
		t.Fatal("expected parse failure for invalid probe format")
	}
}

func TestParseHealthProbesRejectsBadType(t *testing.T) {
	_, err := parseHealthProbes([]string{"ftp:example.com:21"})
	if err == nil {
		t.Fatal("expected parse failure for unsupported type")
	}
}

func TestValidatePolicyWarnsOnMissingBackupCommand(t *testing.T) {
	p := ops.ProjectPolicy{RequireBackup: true, BackupCommand: ""}
	warnings := validatePolicy(p)
	if len(warnings) == 0 {
		t.Fatal("expected warning for require-backup without backup-command")
	}
}

func TestValidatePolicyWarnsOnMissingRestoreCommand(t *testing.T) {
	p := ops.ProjectPolicy{BackupCommand: "pg_dump", RestoreCommand: ""}
	warnings := validatePolicy(p)
	if len(warnings) == 0 {
		t.Fatal("expected warning for backup-command without restore-command")
	}
}

func TestValidatePolicyNoWarningsWhenComplete(t *testing.T) {
	p := ops.ProjectPolicy{RequireBackup: true, BackupCommand: "pg_dump", RestoreCommand: "pg_restore"}
	warnings := validatePolicy(p)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
}

func TestPlanRenderTypeCompatibility(t *testing.T) {
	probe := ops.HealthProbe{Type: "http", Target: "http://localhost", Timeout: 5}
	if probe.Type != "http" {
		t.Fatalf("unexpected probe type %s", probe.Type)
	}
}
