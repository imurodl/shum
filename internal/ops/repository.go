package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/your-org/shum/internal/store"
)

// DBTX is the common interface satisfied by both *sql.DB and *sql.Tx,
// allowing repository methods to operate within or outside a transaction.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Repository struct {
	db       *store.Store
	executor DBTX
}

func NewRepository(db *store.Store) *Repository {
	return &Repository{db: db, executor: db.DB()}
}

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{db: r.db, executor: tx}
}

func (r *Repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx)
}

func (r *Repository) UpsertPolicy(ctx context.Context, p ProjectPolicy) error {
	healthJSON, err := p.ProbeJSON()
	if err != nil {
		return err
	}
	query := `
INSERT INTO project_policies (
	host_alias, project_ref, require_backup, backup_command, restore_command, health_checks, migration_warning
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(host_alias, project_ref) DO UPDATE SET
	require_backup=excluded.require_backup,
	backup_command=excluded.backup_command,
	restore_command=excluded.restore_command,
	health_checks=excluded.health_checks,
	migration_warning=excluded.migration_warning,
	updated_at=strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
`
	_, err = r.executor.ExecContext(
		ctx,
		query,
		p.HostAlias,
		p.ProjectRef,
		boolToInt(p.RequireBackup),
		p.BackupCommand,
		p.RestoreCommand,
		healthJSON,
		boolToInt(p.MigrationWarning),
	)
	return err
}

func (r *Repository) GetPolicy(ctx context.Context, hostAlias, projectRef string) (ProjectPolicy, error) {
	var p ProjectPolicy
	var rawHealth string
	var requireBackup int
	var migrationWarning int
	row := r.executor.QueryRowContext(ctx, `
		SELECT host_alias, project_ref, require_backup, backup_command, restore_command, health_checks, migration_warning
		FROM project_policies
		WHERE host_alias = ? AND project_ref = ?
	`, hostAlias, projectRef)

	if err := row.Scan(&p.HostAlias, &p.ProjectRef, &requireBackup, &p.BackupCommand, &p.RestoreCommand, &rawHealth, &migrationWarning); err != nil {
		if err == sql.ErrNoRows {
			return defaultPolicy(hostAlias, projectRef), nil
		}
		return ProjectPolicy{}, err
	}
	p.RequireBackup = requireBackup == 1
	p.MigrationWarning = migrationWarning == 1
	p.HealthChecks = ParseProbeConfig(rawHealth)
	if p.HealthChecks == nil {
		p.HealthChecks = []HealthProbe{}
	}
	if p.BackupCommand == "" {
		p.BackupCommand = ""
	}
	return p, nil
}

func (r *Repository) SetDefaultPolicy(ctx context.Context, hostAlias, projectRef string) error {
	p := defaultPolicy(hostAlias, projectRef)
	return r.UpsertPolicy(ctx, p)
}

func defaultPolicy(hostAlias, projectRef string) ProjectPolicy {
	return ProjectPolicy{
		HostAlias:        hostAlias,
		ProjectRef:       projectRef,
		RequireBackup:    true,
		BackupCommand:    "",
		RestoreCommand:   "",
		HealthChecks:     []HealthProbe{},
		MigrationWarning: false,
	}
}

func (r *Repository) ListBackups(ctx context.Context, hostAlias, projectRef string) ([]BackupResult, error) {
	rows, err := r.executor.QueryContext(ctx, `
		SELECT id, host_alias, project_ref, artifact_path, artifact_sha256, command, created_at
		FROM backups
		WHERE host_alias = ? AND project_ref = ?
		ORDER BY created_at DESC`, hostAlias, projectRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]BackupResult, 0, 16)
	for rows.Next() {
		var b BackupResult
		var createdRaw string
		if err := rows.Scan(&b.ID, &b.HostAlias, &b.ProjectRef, &b.ArtifactPath, &b.ArtifactSHA, &b.Command, &createdRaw); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339, createdRaw)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) RecordBackup(ctx context.Context, hostAlias, projectRef, artifactPath, artifactSHA, command string) (BackupResult, error) {
	var id int64
	var createdRaw string
	result, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO backups (host_alias, project_ref, artifact_path, artifact_sha256, command)
		 VALUES (?, ?, ?, ?, ?)`,
		hostAlias, projectRef, artifactPath, artifactSHA, command,
	)
	if err != nil {
		return BackupResult{}, err
	}
	id, err = result.LastInsertId()
	if err != nil {
		return BackupResult{}, err
	}
	row := r.executor.QueryRowContext(ctx, `
		SELECT created_at
		FROM backups
		WHERE id = ?`, id)
	if err := row.Scan(&createdRaw); err != nil {
		return BackupResult{}, err
	}
	createdAt, _ := time.Parse(time.RFC3339, createdRaw)
	return BackupResult{
		ID:           id,
		HostAlias:    hostAlias,
		ProjectRef:   projectRef,
		ArtifactPath: artifactPath,
		ArtifactSHA:  artifactSHA,
		Command:      command,
		CreatedAt:    createdAt,
	}, nil
}

func (r *Repository) CreateRun(ctx context.Context, runID, hostAlias, projectRef, planJSON string, preflight *PreflightResult) (RunRecord, error) {
	createdRaw := time.Now().UTC().Format(time.RFC3339)
	preflightJSON := "{}"
	if preflight != nil {
		raw, err := json.Marshal(preflight)
		if err != nil {
			return RunRecord{}, err
		}
		preflightJSON = string(raw)
	}
	if _, err := r.executor.ExecContext(
		ctx,
		`INSERT INTO upgrade_runs (run_id, host_alias, project_ref, status, started_at, preflight, plan)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		runID,
		hostAlias,
		projectRef,
		RunStatusPlanned,
		createdRaw,
		preflightJSON,
		planJSON,
	); err != nil {
		return RunRecord{}, err
	}
	id := int64(0)
	row := r.executor.QueryRowContext(
		ctx,
		`SELECT id FROM upgrade_runs WHERE run_id = ?`,
		runID,
	)
	if err := row.Scan(&id); err != nil {
		return RunRecord{}, err
	}
	return RunRecord{
		ID:         id,
		RunID:      runID,
		HostAlias:  hostAlias,
		ProjectRef: projectRef,
		Status:     RunStatusPlanned,
		StartedAt:  time.Now().UTC(),
		Preflight:  preflightJSON,
		Plan:       planJSON,
	}, nil
}

func (r *Repository) UpdateRun(ctx context.Context, runID string, status RunStatus, summary, failureReason, backupArtifact string, finished bool) error {
	if err := validateRunStatus(status); err != nil {
		return err
	}
	now := ""
	if finished {
		now = time.Now().UTC().Format(time.RFC3339)
	}
	query := `UPDATE upgrade_runs SET status = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`
	args := []any{status}
	if summary != "" {
		query += `, summary = ?`
		args = append(args, summary)
	}
	if failureReason != "" {
		query += `, failure_reason = ?`
		args = append(args, failureReason)
	}
	if backupArtifact != "" {
		query += `, backup_artifact = ?`
		args = append(args, backupArtifact)
	}
	if finished {
		query += `, finished_at = ?`
		args = append(args, now)
	}
	query += ` WHERE run_id = ?`
	args = append(args, runID)
	_, err := r.executor.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) AddRunEvent(ctx context.Context, runID, eventType, detail string) error {
	_, err := r.executor.ExecContext(ctx,
		`INSERT INTO upgrade_run_events (run_id, event_type, detail) VALUES (?, ?, ?)`,
		runID, eventType, detail,
	)
	return err
}

func (r *Repository) GetRun(ctx context.Context, runID string) (RunRecord, error) {
	var out RunRecord
	var preflightRaw, planRaw, status string
	var startedRaw, finishedRaw sql.NullString
	var summary, backupArtifact, failureReason sql.NullString
	row := r.executor.QueryRowContext(
		ctx,
		`SELECT id, run_id, host_alias, project_ref, status, started_at, finished_at, preflight, plan, summary, backup_artifact, failure_reason
		 FROM upgrade_runs WHERE run_id = ?`, runID)
	if err := row.Scan(&out.ID, &out.RunID, &out.HostAlias, &out.ProjectRef, &status, &startedRaw, &finishedRaw, &preflightRaw, &planRaw, &summary, &backupArtifact, &failureReason); err != nil {
		return RunRecord{}, err
	}
	out.Status = RunStatus(status)
	if startedRaw.Valid {
		out.StartedAt, _ = time.Parse(time.RFC3339, startedRaw.String)
	}
	if finishedRaw.Valid {
		out.FinishedAt, _ = time.Parse(time.RFC3339, finishedRaw.String)
	}
	out.Preflight = preflightRaw
	out.Plan = planRaw
	if summary.Valid {
		out.Summary = summary.String
	}
	if backupArtifact.Valid {
		out.BackupArtifact = backupArtifact.String
	}
	if failureReason.Valid {
		out.FailureReason = failureReason.String
	}
	events, err := r.listEvents(ctx, runID)
	if err != nil {
		return RunRecord{}, err
	}
	out.Events = events
	return out, nil
}

func (r *Repository) listEvents(ctx context.Context, runID string) ([]RunEvent, error) {
	rows, err := r.executor.QueryContext(ctx, `
		SELECT id, run_id, event_type, detail, created_at
		FROM upgrade_run_events
		WHERE run_id = ?
		ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RunEvent, 0, 16)
	for rows.Next() {
		var e RunEvent
		var createdRaw string
		if err := rows.Scan(&e.ID, &e.RunID, &e.Type, &e.Detail, &createdRaw); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdRaw)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) ListRuns(ctx context.Context, limit int, hostAlias, projectRef string) ([]RunRecord, error) {
	stmt := `SELECT id, run_id, host_alias, project_ref, status, started_at, finished_at, summary, failure_reason
		FROM upgrade_runs
		WHERE 1=1`
	args := make([]any, 0, 2)
	if hostAlias != "" {
		stmt += ` AND host_alias = ?`
		args = append(args, hostAlias)
	}
	if projectRef != "" {
		stmt += ` AND project_ref = ?`
		args = append(args, projectRef)
	}
	stmt += ` ORDER BY created_at DESC`
	if limit > 0 {
		stmt += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := r.executor.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RunRecord, 0, 16)
	for rows.Next() {
		var rrow RunRecord
		var status string
		var startedRaw, finishedRaw sql.NullString
		var summary, failure sql.NullString
		if err := rows.Scan(&rrow.ID, &rrow.RunID, &rrow.HostAlias, &rrow.ProjectRef, &status, &startedRaw, &finishedRaw, &summary, &failure); err != nil {
			return nil, err
		}
		rrow.Status = RunStatus(status)
		if startedRaw.Valid {
			rrow.StartedAt, _ = time.Parse(time.RFC3339, startedRaw.String)
		}
		if finishedRaw.Valid {
			rrow.FinishedAt, _ = time.Parse(time.RFC3339, finishedRaw.String)
		}
		if summary.Valid {
			rrow.Summary = summary.String
		}
		if failure.Valid {
			rrow.FailureReason = failure.String
		}
		out = append(out, rrow)
	}
	return out, rows.Err()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
