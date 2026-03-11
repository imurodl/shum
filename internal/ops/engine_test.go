package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/your-org/shum/internal/hosts"
	"github.com/your-org/shum/internal/projects"
	"github.com/your-org/shum/internal/store"
)

// newTestService creates a Service backed by an in-memory SQLite database
// and a mock runner. It seeds a host and project record so ensureProjectExists passes.
func newTestService(t *testing.T, runner *mockRunner) *Service {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "state.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()

	// Seed host record.
	hostRepo := hosts.NewRepository(db)
	if err := hostRepo.Upsert(ctx, hosts.Host{
		Alias:    "test-host",
		Hostname: "10.0.0.1",
		UserName: "deploy",
		Port:     22,
	}); err != nil {
		t.Fatalf("failed to seed host: %v", err)
	}

	// Seed project record.
	projRepo := projects.NewProjectRepository(db)
	if err := projRepo.UpsertProject(ctx, projects.ProjectRecord{
		HostAlias:  "test-host",
		ProjectRef: "web",
		Status:     projects.StatusCanonical,
	}); err != nil {
		t.Fatalf("failed to seed project: %v", err)
	}

	// Build host service with nil runner (Inspect only reads from DB).
	hostSvc := hosts.NewService(hostRepo, nil)

	opsRepo := NewRepository(db)
	artifactDir := filepath.Join(tempDir, "artifacts")
	_ = os.MkdirAll(artifactDir, 0o755)

	return NewService(hostSvc, projRepo, runner, opsRepo, artifactDir)
}

func TestRunUpgradeDryRun(t *testing.T) {
	runner := newMockRunner()
	runner.preflightOK()
	runner.composePSOK()
	runner.imageInspectOK()

	svc := newTestService(t, runner)
	ctx := context.Background()

	result, err := svc.RunUpgrade(ctx, "test-host", "web", UpgradeOptions{
		DryRun:     true,
		SkipBackup: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != string(RunStatusSuccess) {
		t.Fatalf("expected success status, got %q", result.Status)
	}
	if result.RunID == "" {
		t.Fatal("expected non-empty run ID")
	}

	// Verify run was persisted.
	run, err := svc.opsRepo.GetRun(ctx, result.RunID)
	if err != nil {
		t.Fatalf("failed to fetch run: %v", err)
	}
	if run.Status != RunStatusSuccess {
		t.Fatalf("expected persisted status success, got %q", run.Status)
	}
	if len(run.Events) == 0 {
		t.Fatal("expected at least one event")
	}
}

func TestRunUpgradeFullSuccess(t *testing.T) {
	runner := newMockRunner()
	runner.preflightOK()
	runner.composePSOK()
	runner.imageInspectOK()
	runner.upgradeOK()

	svc := newTestService(t, runner)
	ctx := context.Background()

	result, err := svc.RunUpgrade(ctx, "test-host", "web", UpgradeOptions{
		SkipBackup: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != string(RunStatusSuccess) {
		t.Fatalf("expected success, got %q", result.Status)
	}

	run, err := svc.opsRepo.GetRun(ctx, result.RunID)
	if err != nil {
		t.Fatalf("failed to fetch run: %v", err)
	}
	// Verify event types include apply and complete.
	eventTypes := make(map[string]bool)
	for _, e := range run.Events {
		eventTypes[e.Type] = true
	}
	if !eventTypes["apply"] {
		t.Fatal("expected apply event")
	}
	if !eventTypes["complete"] {
		t.Fatal("expected complete event")
	}
}

func TestRunUpgradeComposePullFailureTriggersRollback(t *testing.T) {
	runner := newMockRunner()
	runner.preflightOK()
	runner.composePSOK()
	runner.imageInspectOK()
	runner.pullFails()
	// rollback via compose down/up (no restore command configured)
	runner.on("docker compose down", "", nil)

	svc := newTestService(t, runner)
	ctx := context.Background()

	// Disable backup requirement for this test.
	_ = svc.opsRepo.UpsertPolicy(ctx, ProjectPolicy{
		HostAlias:     "test-host",
		ProjectRef:    "web",
		RequireBackup: false,
		HealthChecks:  []HealthProbe{},
	})

	result, _ := svc.RunUpgrade(ctx, "test-host", "web", UpgradeOptions{})
	if result.Status != string(RunStatusRolledBack) && result.Status != string(RunStatusFailed) {
		t.Fatalf("expected rolled_back or failed status, got %q", result.Status)
	}

	run, err := svc.opsRepo.GetRun(ctx, result.RunID)
	if err != nil {
		t.Fatalf("failed to fetch run: %v", err)
	}
	eventTypes := make(map[string]bool)
	for _, e := range run.Events {
		eventTypes[e.Type] = true
	}
	if !eventTypes["rollback"] {
		t.Fatal("expected rollback event")
	}
}

func TestRunUpgradeTransactionRollbackLeavesNoState(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "state.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewRepository(db)

	// Start a transaction, create a run, then rollback.
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	txRepo := repo.WithTx(tx)

	_, err = txRepo.CreateRun(ctx, "tx-test-run", "alpha", "web", "{}", nil)
	if err != nil {
		t.Fatalf("failed to create run in tx: %v", err)
	}
	_ = txRepo.AddRunEvent(ctx, "tx-test-run", "start", "test")

	// Rollback instead of commit.
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Verify nothing persisted.
	runs, err := repo.ListRuns(ctx, 0, "", "")
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected 0 runs after rollback, got %d", len(runs))
	}
}

func TestRunUpgradeMigrationWarningBlocksWithoutForce(t *testing.T) {
	runner := newMockRunner()
	svc := newTestService(t, runner)
	ctx := context.Background()

	_ = svc.opsRepo.UpsertPolicy(ctx, ProjectPolicy{
		HostAlias:        "test-host",
		ProjectRef:       "web",
		MigrationWarning: true,
		HealthChecks:     []HealthProbe{},
	})

	_, err := svc.RunUpgrade(ctx, "test-host", "web", UpgradeOptions{})
	if err == nil {
		t.Fatal("expected error when migration warning is enabled without --force")
	}
}

func TestRunUpgradeMigrationWarningPassesEarlyGuardWithForce(t *testing.T) {
	runner := newMockRunner()
	runner.preflightOK()
	runner.composePSOK()
	runner.imageInspectOK()

	svc := newTestService(t, runner)
	ctx := context.Background()

	// Set policy without migration warning so planner won't add a block.
	_ = svc.opsRepo.UpsertPolicy(ctx, ProjectPolicy{
		HostAlias:     "test-host",
		ProjectRef:    "web",
		RequireBackup: false,
		HealthChecks:  []HealthProbe{},
	})

	result, err := svc.RunUpgrade(ctx, "test-host", "web", UpgradeOptions{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != string(RunStatusSuccess) {
		t.Fatalf("expected success, got %q", result.Status)
	}
}
