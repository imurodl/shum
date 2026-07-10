package projects

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/imurodl/shum/internal/shumerr"
	"github.com/imurodl/shum/internal/store"
)

func newTestRepo(t *testing.T) *ProjectRepository {
	t.Helper()
	s, err := store.New(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return NewProjectRepository(s)
}

func TestUpsertAndGetProject(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	want := ProjectRecord{
		HostAlias:        "prod",
		ProjectRef:       "web",
		Status:           StatusCanonical,
		Canonical:        true,
		ProjectName:      "web",
		ProjectDirectory: "/srv/web",
		ComposeFiles:     []string{"docker-compose.yml", "docker-compose.prod.yml"},
		ActiveProfiles:   []string{"default"},
		EnvFingerprint:   "abc123",
	}
	if err := repo.UpsertProject(ctx, want); err != nil {
		t.Fatalf("UpsertProject: %v", err)
	}

	got, err := repo.GetProject(ctx, "prod", "web")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Status != want.Status || !got.Canonical {
		t.Errorf("status/canonical = %q/%v, want %q/true", got.Status, got.Canonical, want.Status)
	}
	if got.ProjectDirectory != want.ProjectDirectory || got.EnvFingerprint != want.EnvFingerprint {
		t.Errorf("scalar fields did not round-trip: %+v", got)
	}
	if len(got.ComposeFiles) != 2 || got.ComposeFiles[1] != "docker-compose.prod.yml" {
		t.Errorf("ComposeFiles did not round-trip: %v", got.ComposeFiles)
	}
	if len(got.ActiveProfiles) != 1 || got.ActiveProfiles[0] != "default" {
		t.Errorf("ActiveProfiles did not round-trip: %v", got.ActiveProfiles)
	}
}

func TestUpsertProjectConflictUpdates(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	base := ProjectRecord{HostAlias: "prod", ProjectRef: "web", Status: StatusRuntimeOnly}
	if err := repo.UpsertProject(ctx, base); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	base.Status = StatusCanonical
	base.Canonical = true
	if err := repo.UpsertProject(ctx, base); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	list, err := repo.ListByHost(ctx, "prod")
	if err != nil {
		t.Fatalf("ListByHost: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ON CONFLICT should update in place, got %d rows", len(list))
	}
	if list[0].Status != StatusCanonical || !list[0].Canonical {
		t.Errorf("conflict update not applied: %+v", list[0])
	}
}

func TestGetProjectNotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.GetProject(context.Background(), "prod", "missing")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
	se, ok := shumerr.From(err)
	if !ok {
		t.Fatalf("expected a shumerr.Error, got %T: %v", err, err)
	}
	if se.Code != shumerr.CodeProjectNotFound {
		t.Errorf("code = %q, want %q", se.Code, shumerr.CodeProjectNotFound)
	}
}

func TestListByHostEmpty(t *testing.T) {
	repo := newTestRepo(t)
	list, err := repo.ListByHost(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("ListByHost: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected no rows, got %d", len(list))
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 || boolToInt(false) != 0 {
		t.Error("boolToInt mapping is wrong")
	}
}
