package projects

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imurodl/shum/internal/store"
)

type ProjectRecord struct {
	ID               int64
	HostAlias        string
	ProjectRef       string
	Status           ProjectStatus
	Canonical        bool
	ProjectName      string
	ProjectDirectory string
	ComposeFiles     []string
	ActiveProfiles   []string
	EnvFingerprint   string
	DiscoveredAt     time.Time
	UpdatedAt        time.Time
}

type ProjectRepository struct {
	db *store.Store
}

type DiscoverySnapshot struct {
	ID        int64
	HostAlias string
	Payload   map[string]any
	CreatedAt time.Time
}

func NewProjectRepository(db *store.Store) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) UpsertProject(ctx context.Context, p ProjectRecord) error {
	filesJSON, err := json.Marshal(p.ComposeFiles)
	if err != nil {
		return err
	}
	profilesJSON, err := json.Marshal(p.ActiveProfiles)
	if err != nil {
		return err
	}
	_, err = r.db.DB().ExecContext(
		ctx,
		`INSERT INTO projects (
			host_alias, project_ref, status, canonical, project_name, project_directory, compose_files, active_profiles, env_fingerprint
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(host_alias, project_ref) DO UPDATE SET
			status=excluded.status,
			canonical=excluded.canonical,
			project_name=excluded.project_name,
			project_directory=excluded.project_directory,
			compose_files=excluded.compose_files,
			active_profiles=excluded.active_profiles,
			env_fingerprint=excluded.env_fingerprint,
			updated_at=strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`,
		p.HostAlias,
		p.ProjectRef,
		string(p.Status),
		boolToInt(p.Canonical),
		p.ProjectName,
		p.ProjectDirectory,
		string(filesJSON),
		string(profilesJSON),
		p.EnvFingerprint,
	)
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
} 
	return 0
}

func (r *ProjectRepository) ListByHost(ctx context.Context, hostAlias string) ([]ProjectRecord, error) {
	rows, err := r.db.DB().QueryContext(
		ctx,
		`SELECT id, project_ref, status, canonical, project_name, project_directory, compose_files, active_profiles, env_fingerprint, discovered_at, updated_at
		 FROM projects WHERE host_alias = ? ORDER BY project_ref`,
		hostAlias,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProjectRecord, 0, 16)
	for rows.Next() {
		var p ProjectRecord
		var status, filesRaw, profileRaw, discoveredRaw, updatedRaw string
		if err := rows.Scan(
			&p.ID,
			&p.ProjectRef,
			&status,
			&p.Canonical,
			&p.ProjectName,
			&p.ProjectDirectory,
			&filesRaw,
			&profileRaw,
			&p.EnvFingerprint,
			&discoveredRaw,
			&updatedRaw,
		); err != nil {
			return nil, err
		}
		p.HostAlias = hostAlias
		p.Status = ProjectStatus(status)
		_ = json.Unmarshal([]byte(filesRaw), &p.ComposeFiles)
		_ = json.Unmarshal([]byte(profileRaw), &p.ActiveProfiles)
		p.DiscoveredAt, _ = time.Parse(time.RFC3339, discoveredRaw)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedRaw)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProjectRepository) GetProject(ctx context.Context, hostAlias, ref string) (ProjectRecord, error) {
	var p ProjectRecord
	var status, filesRaw, profileRaw, discoveredRaw, updatedRaw string
	row := r.db.DB().QueryRowContext(
		ctx,
		`SELECT id, status, canonical, project_name, project_directory, compose_files, active_profiles, env_fingerprint, discovered_at, updated_at
		 FROM projects WHERE host_alias = ? AND project_ref = ?`,
		hostAlias, ref,
	)
	err := row.Scan(
		&p.ID,
		&status,
		&p.Canonical,
		&p.ProjectName,
		&p.ProjectDirectory,
		&filesRaw,
		&profileRaw,
		&p.EnvFingerprint,
		&discoveredRaw,
		&updatedRaw,
	)
	if err == sql.ErrNoRows {
		return ProjectRecord{}, fmt.Errorf("project not found")
	}
	if err != nil {
		return ProjectRecord{}, err
	}
	p.HostAlias = hostAlias
	p.ProjectRef = ref
	p.Status = ProjectStatus(status)
	_ = json.Unmarshal([]byte(filesRaw), &p.ComposeFiles)
	_ = json.Unmarshal([]byte(profileRaw), &p.ActiveProfiles)
	p.DiscoveredAt, _ = time.Parse(time.RFC3339, discoveredRaw)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedRaw)
	return p, nil
}
