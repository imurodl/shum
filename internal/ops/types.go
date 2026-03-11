package ops

import (
	"encoding/json"
	"fmt"
	"time"
)

type HealthProbe struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Timeout int    `json:"timeout_seconds"`
}

type ProjectPolicy struct {
	HostAlias        string       `json:"-"`
	ProjectRef       string       `json:"-"`
	RequireBackup    bool         `json:"require_backup"`
	BackupCommand    string       `json:"backup_command"`
	RestoreCommand   string       `json:"restore_command"`
	HealthChecks     []HealthProbe `json:"health_checks"`
	MigrationWarning bool         `json:"migration_warning"`
}

func (p ProjectPolicy) ProbeJSON() (string, error) {
	raw, err := json.Marshal(p.HealthChecks)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func ParseProbeConfig(raw string) []HealthProbe {
	if raw == "" {
		return []HealthProbe{}
	}
	var out []HealthProbe
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []HealthProbe{}
	}
	return out
}

type PreflightResult struct {
	HostAlias        string            `json:"host_alias"`
	DockerAvailable  bool              `json:"docker_available"`
	ComposeAvailable bool              `json:"compose_available"`
	DockerVersion    string            `json:"docker_version"`
	ComposeVersion   string            `json:"compose_version"`
	DiskBytesAvail   int64             `json:"disk_bytes_available"`
	DiskPath         string            `json:"disk_path"`
	PermissionsOK    bool              `json:"permissions_ok"`
	Checks           map[string]string `json:"checks"`
	Passed           bool              `json:"passed"`
}

type ServiceChange struct {
	ServiceName    string `json:"service"`
	Image          string `json:"image"`
	CurrentDigest  string `json:"current_digest"`
	TargetDigest   string `json:"target_digest"`
	CanRollback    bool   `json:"can_rollback"`
	CurrentHealthy string `json:"current_health"`
}

type Plan struct {
	HostAlias  string                  `json:"host_alias"`
	ProjectRef string                  `json:"project_ref"`
	RunID      string                  `json:"run_id"`
	Preflight  PreflightResult         `json:"preflight"`
	Services   []ServiceChange         `json:"services"`
	Actions    []PlanAction            `json:"actions"`
	Policy     ProjectPolicy           `json:"policy"`
	CreatedAt  string                  `json:"created_at"`
	Warnings   []string                `json:"warnings"`
	Blocks     []string                `json:"blocks"`
}

type PlanAction struct {
	Name  string `json:"name"`
	Cmd   string `json:"cmd"`
	Notes string `json:"notes"`
}

func (p *Plan) String() string {
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "invalid plan"
	}
	return string(raw)
}

type BackupResult struct {
	ID            int64     `json:"id"`
	HostAlias     string    `json:"host_alias"`
	ProjectRef    string    `json:"project_ref"`
	ArtifactPath  string    `json:"artifact_path"`
	ArtifactSHA   string    `json:"artifact_sha256"`
	Command       string    `json:"command"`
	CreatedAt     time.Time `json:"created_at"`
	RecordedBytes int64     `json:"size_bytes"`
}

type RunStatus string

const (
	RunStatusPlanned   RunStatus = "planned"
	RunStatusRunning   RunStatus = "running"
	RunStatusSuccess   RunStatus = "success"
	RunStatusFailed    RunStatus = "failed"
	RunStatusRolledBack RunStatus = "rolled_back"
)

type RunRecord struct {
	ID             int64       `json:"id"`
	RunID          string      `json:"run_id"`
	HostAlias      string      `json:"host_alias"`
	ProjectRef     string      `json:"project_ref"`
	Status         RunStatus   `json:"status"`
	StartedAt      time.Time   `json:"started_at"`
	FinishedAt     time.Time   `json:"finished_at"`
	Preflight      string      `json:"preflight"`
	Plan           string      `json:"plan"`
	Summary        string      `json:"summary"`
	BackupArtifact string      `json:"backup_artifact"`
	FailureReason  string      `json:"failure_reason"`
	Events         []RunEvent  `json:"events"`
}

type RunEvent struct {
	ID        int64     `json:"id"`
	RunID     string    `json:"run_id"`
	Type      string    `json:"type"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

func validateRunStatus(status RunStatus) error {
	switch status {
	case RunStatusPlanned, RunStatusRunning, RunStatusSuccess, RunStatusFailed, RunStatusRolledBack:
		return nil
	default:
		return fmt.Errorf("invalid run status: %s", status)
	}
}
