package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/imurodl/shum/internal/store"
)

func TestListRunsFiltersByHostAndProject(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "shum-repo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	dbPath := filepath.Join(tempDir, "state.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	repo := NewRepository(db)

	_, err = repo.CreateRun(ctx, "run-a1", "alpha", "web", "{}", nil)
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}
	_, err = repo.CreateRun(ctx, "run-a2", "alpha", "api", "{}", nil)
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}
	_, err = repo.CreateRun(ctx, "run-b1", "beta", "web", "{}", nil)
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	allRuns, err := repo.ListRuns(ctx, 0, "", "")
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(allRuns) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(allRuns))
	}

	alphaRuns, err := repo.ListRuns(ctx, 0, "alpha", "")
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(alphaRuns) != 2 {
		t.Fatalf("expected 2 runs for alpha host, got %d", len(alphaRuns))
	}
	for _, run := range alphaRuns {
		if run.HostAlias != "alpha" {
			t.Fatalf("expected host alpha in filtered set, got %q", run.HostAlias)
		}
	}

	alphaWebRuns, err := repo.ListRuns(ctx, 0, "alpha", "web")
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(alphaWebRuns) != 1 {
		t.Fatalf("expected 1 run for alpha web, got %d", len(alphaWebRuns))
	}
	if alphaWebRuns[0].ProjectRef != "web" {
		t.Fatalf("expected web project in filtered run, got %q", alphaWebRuns[0].ProjectRef)
	}

	limited, err := repo.ListRuns(ctx, 1, "alpha", "")
	if err != nil {
		t.Fatalf("list runs failed: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("expected 1 limited run, got %d", len(limited))
	}
}
