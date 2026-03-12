package ops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imurodl/shum/internal/hosts"
	"github.com/imurodl/shum/internal/projects"
)

type Service struct {
	hostService *hosts.Service
	projectRepo *projects.ProjectRepository
	opsRepo     *Repository
	planner     *PlanningService
	runner      remoteRunner
	artifactDir string
}

type UpgradeOptions struct {
	Force      bool
	SkipBackup bool
	DryRun     bool
	HttpProbes []string
	TcpProbes  []string
	CmdProbes  []string
}

type UpgradeResult struct {
	RunID   string `json:"run_id"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

func NewService(hostService *hosts.Service, projectRepo *projects.ProjectRepository, runner remoteRunner, opsRepo *Repository, artifactDir string) *Service {
	return &Service{
		hostService: hostService,
		projectRepo: projectRepo,
		opsRepo:     opsRepo,
		planner:     NewPlanningService(runner),
		runner:      runner,
		artifactDir: artifactDir,
	}
}

func (s *Service) Preflight(ctx context.Context, hostAlias, projectRef string) (PreflightResult, error) {
	if err := s.ensureProjectExists(ctx, hostAlias, projectRef); err != nil {
		return PreflightResult{}, err
	}
	return s.planner.Preflight(ctx, hostAlias)
}

func (s *Service) ResolvePolicy(ctx context.Context, hostAlias, projectRef string) (ProjectPolicy, error) {
	return s.opsRepo.GetPolicy(ctx, hostAlias, projectRef)
}

func (s *Service) Plan(ctx context.Context, hostAlias, projectRef string, policy *ProjectPolicy) (Plan, error) {
	if err := s.ensureProjectExists(ctx, hostAlias, projectRef); err != nil {
		return Plan{}, err
	}
	if policy == nil {
		p, err := s.opsRepo.GetPolicy(ctx, hostAlias, projectRef)
		if err != nil {
			return Plan{}, err
		}
		policy = &p
	}
	return s.planner.BuildPlan(ctx, hostAlias, projectRef, *policy)
}

func (s *Service) SetPolicy(ctx context.Context, p ProjectPolicy) error {
	if p.ProjectRef == "" || p.HostAlias == "" {
		return fmt.Errorf("host alias and project ref required")
	}
	return s.opsRepo.UpsertPolicy(ctx, p)
}

func (s *Service) ListBackups(ctx context.Context, hostAlias, projectRef string) ([]BackupResult, error) {
	if err := s.ensureProjectExists(ctx, hostAlias, projectRef); err != nil {
		return nil, err
	}
	return s.opsRepo.ListBackups(ctx, hostAlias, projectRef)
}

func (s *Service) TakeBackup(ctx context.Context, hostAlias, projectRef string, command string) (BackupResult, error) {
	if err := s.ensureProjectExists(ctx, hostAlias, projectRef); err != nil {
		return BackupResult{}, err
	}
	policy, err := s.opsRepo.GetPolicy(ctx, hostAlias, projectRef)
	if err != nil {
		return BackupResult{}, err
	}

	artifactPath := filepath.Join(s.artifactDir, "backups", hostAlias, projectRef, fmt.Sprintf("%d.txt", time.Now().UnixNano()))
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		return BackupResult{}, err
	}

	cmd := command
	if strings.TrimSpace(cmd) == "" {
		cmd = strings.TrimSpace(policy.BackupCommand)
	}
	if cmd == "" {
		return BackupResult{}, fmt.Errorf("no backup command configured")
	}

	// Start with an empty artifact so restore commands that write to path are supported.
	if err := os.WriteFile(artifactPath, []byte{}, 0o600); err != nil {
		return BackupResult{}, err
	}

	enriched := fmt.Sprintf("SHUM_BACKUP_ARTIFACT=%s %s", shellEscape(artifactPath), cmd)
	out, err := s.runner.Command(hostAlias, enriched)
	if err != nil {
		return BackupResult{}, fmt.Errorf("backup command failed: %w: %s", err, out)
	}

	if output := strings.TrimSpace(out); output != "" {
		if err := os.WriteFile(artifactPath, []byte(output), 0o600); err != nil {
			return BackupResult{}, err
		}
	}

	artifactBytes, err := os.ReadFile(artifactPath)
	if err != nil {
		return BackupResult{}, err
	}
	sum := sha256.Sum256(artifactBytes)
	record, err := s.opsRepo.RecordBackup(
		ctx,
		hostAlias,
		projectRef,
		artifactPath,
		hex.EncodeToString(sum[:]),
		enriched,
	)
	if err != nil {
		return BackupResult{}, err
	}
	return record, nil
}

func (s *Service) RestoreBackup(ctx context.Context, hostAlias, projectRef string, artifactPath, command string) error {
	if err := s.ensureProjectExists(ctx, hostAlias, projectRef); err != nil {
		return err
	}
	if strings.TrimSpace(command) == "" {
		p, err := s.opsRepo.GetPolicy(ctx, hostAlias, projectRef)
		if err != nil {
			return err
		}
		command = p.RestoreCommand
	}
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("no restore command configured")
	}

	if _, err := os.Stat(artifactPath); err != nil {
		return fmt.Errorf("artifact not found: %w", err)
	}
	enriched := fmt.Sprintf("SHUM_BACKUP_ARTIFACT=%s %s", shellEscape(artifactPath), command)
	if _, err := s.runner.Command(hostAlias, enriched); err != nil {
		return fmt.Errorf("restore command failed: %w", err)
	}
	return nil
}

func (s *Service) RunUpgrade(ctx context.Context, hostAlias, projectRef string, opts UpgradeOptions) (UpgradeResult, error) {
	if err := s.ensureProjectExists(ctx, hostAlias, projectRef); err != nil {
		return UpgradeResult{}, err
	}
	policy, err := s.opsRepo.GetPolicy(ctx, hostAlias, projectRef)
	if err != nil {
		return UpgradeResult{}, err
	}
	if !opts.Force && policy.MigrationWarning {
		return UpgradeResult{}, fmt.Errorf("migration warning is enabled; use --force to continue")
	}

	plan, err := s.Plan(ctx, hostAlias, projectRef, &policy)
	if err != nil {
		return UpgradeResult{}, err
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return UpgradeResult{}, err
	}

	tx, err := s.opsRepo.BeginTx(ctx)
	if err != nil {
		return UpgradeResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op after commit
	repo := s.opsRepo.WithTx(tx)

	runID := fmt.Sprintf("run-%d", time.Now().UnixNano())
	run, err := repo.CreateRun(ctx, runID, hostAlias, projectRef, string(planJSON), &plan.Preflight)
	if err != nil {
		return UpgradeResult{}, err
	}
	_ = run
	_ = repo.AddRunEvent(ctx, runID, "start", "upgrade started")
	if err := repo.UpdateRun(ctx, runID, RunStatusRunning, "", "", "", false); err != nil {
		return UpgradeResult{}, err
	}
	_ = repo.AddRunEvent(ctx, runID, "plan", plan.String())

	backup, needBackup := BackupResult{}, policy.RequireBackup && !opts.SkipBackup
	if needBackup {
		if backup, err = s.TakeBackup(ctx, hostAlias, projectRef, ""); err != nil {
			_ = repo.UpdateRun(ctx, runID, RunStatusFailed, "", fmt.Sprintf("backup failed: %v", err), "", true)
			_ = repo.AddRunEvent(ctx, runID, "backup", fmt.Sprintf("failed: %v", err))
			_ = tx.Commit()
			return UpgradeResult{RunID: runID, Status: string(RunStatusFailed), Summary: "backup failed"}, err
		}
		_ = repo.UpdateRun(ctx, runID, RunStatusRunning, "", "", backup.ArtifactPath, false)
		_ = repo.AddRunEvent(ctx, runID, "backup", fmt.Sprintf("backup recorded: %s", backup.ArtifactPath))
	}

	if len(plan.Blocks) > 0 {
		blockReason := strings.Join(plan.Blocks, "; ")
		_ = repo.UpdateRun(ctx, runID, RunStatusFailed, blockReason, blockReason, "", true)
		_ = repo.AddRunEvent(ctx, runID, "block", blockReason)
		_ = tx.Commit()
		return UpgradeResult{RunID: runID, Status: string(RunStatusFailed), Summary: blockReason}, fmt.Errorf(blockReason)
	}

	if opts.DryRun {
		msg := "dry-run selected, no mutation executed"
		_ = repo.UpdateRun(ctx, runID, RunStatusSuccess, msg, "", "", true)
		_ = repo.AddRunEvent(ctx, runID, "complete", msg)
		_ = tx.Commit()
		return UpgradeResult{RunID: runID, Status: string(RunStatusSuccess), Summary: msg}, nil
	}

	if _, err := s.runner.Command(hostAlias, "docker compose pull"); err != nil {
		fail := fmt.Sprintf("compose pull failed: %v", err)
		return s.rollbackTx(ctx, tx, repo, runID, hostAlias, projectRef, policy, backup.ArtifactPath, fail)
	}
	_ = repo.AddRunEvent(ctx, runID, "apply", "compose pull completed")

	if _, err := s.runner.Command(hostAlias, "docker compose up -d"); err != nil {
		fail := fmt.Sprintf("compose up failed: %v", err)
		return s.rollbackTx(ctx, tx, repo, runID, hostAlias, projectRef, policy, backup.ArtifactPath, fail)
	}
	_ = repo.AddRunEvent(ctx, runID, "apply", "compose up -d completed")

	probes := mergeProbes(policy.HealthChecks, opts.HttpProbes, opts.TcpProbes, opts.CmdProbes)
	if err := s.verify(ctx, hostAlias, probes, outlierWaitTimeout()); err != nil {
		fail := fmt.Sprintf("health verification failed: %v", err)
		return s.rollbackTx(ctx, tx, repo, runID, hostAlias, projectRef, policy, backup.ArtifactPath, fail)
	}

	if err := repo.UpdateRun(ctx, runID, RunStatusSuccess, "upgrade completed", "", backup.ArtifactPath, true); err != nil {
		return UpgradeResult{}, err
	}
	_ = repo.AddRunEvent(ctx, runID, "complete", "upgrade completed")
	if err := tx.Commit(); err != nil {
		return UpgradeResult{}, fmt.Errorf("failed to commit upgrade record: %w", err)
	}
	return UpgradeResult{RunID: runID, Status: string(RunStatusSuccess), Summary: "upgrade completed"}, nil
}

func (s *Service) rollbackTx(
	ctx context.Context,
	tx interface{ Commit() error },
	repo *Repository,
	runID, hostAlias, projectRef string,
	policy ProjectPolicy,
	artifactPath string,
	reason string,
) (UpgradeResult, error) {
	_ = repo.AddRunEvent(ctx, runID, "rollback", reason)
	var rollbackErr error
	if strings.TrimSpace(policy.RestoreCommand) != "" && artifactPath != "" {
		rollbackErr = s.RestoreBackup(ctx, hostAlias, projectRef, artifactPath, policy.RestoreCommand)
	} else {
		_, rollbackErr = s.runner.Command(hostAlias, "docker compose down && docker compose up -d")
	}
	if rollbackErr != nil {
		statusReason := fmt.Sprintf("rollback failed: %v", rollbackErr)
		_ = repo.UpdateRun(ctx, runID, RunStatusFailed, "", statusReason, artifactPath, true)
		_ = repo.AddRunEvent(ctx, runID, "rollback-failure", statusReason)
		_ = tx.Commit()
		return UpgradeResult{RunID: runID, Status: string(RunStatusFailed), Summary: statusReason}, rollbackErr
	}
	_ = repo.UpdateRun(ctx, runID, RunStatusRolledBack, "upgrade rolled back", reason, artifactPath, true)
	_ = repo.AddRunEvent(ctx, runID, "rollback", "upgrade rolled back")
	_ = tx.Commit()
	return UpgradeResult{RunID: runID, Status: string(RunStatusRolledBack), Summary: reason}, nil
}

func (s *Service) verify(ctx context.Context, hostAlias string, probes []HealthProbe, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(probes) == 0 {
		raw, err := s.runner.Command(hostAlias, "docker compose ps --format json")
		if err != nil {
			return err
		}
		services := extractServiceChanges(raw)
		for _, item := range services {
			select {
			case <-timeoutCtx.Done():
				return timeoutCtx.Err()
			default:
			}
			if !isServiceHealthy(item.CurrentHealthy) {
				return fmt.Errorf("service %s is not running (%s)", item.ServiceName, item.CurrentHealthy)
			}
		}
		return nil
	}

	for _, probe := range probes {
		select {
		case <-timeoutCtx.Done():
			return timeoutCtx.Err()
		default:
		}
		switch probe.Type {
		case "http":
			if err := verifyHTTPProbe(timeoutCtx, probe.Target, probeTimeout(probe.Timeout)); err != nil {
				return err
			}
		case "tcp":
			cmd := fmt.Sprintf("bash -lc 'command -v nc >/dev/null 2>&1 && nc -z %s 2>/dev/null || command -v socat >/dev/null 2>&1 && socat -T 5 /dev/null TCP:%s || exit 1'", probe.Target, probe.Target)
			if _, err := s.runner.Command(hostAlias, cmd); err != nil {
				return fmt.Errorf("tcp probe failed for %s: %w", probe.Target, err)
			}
		case "cmd":
			if _, err := s.runner.Command(hostAlias, probe.Target); err != nil {
				return fmt.Errorf("cmd probe failed for %s: %w", probe.Target, err)
			}
		default:
			return fmt.Errorf("unknown probe type: %s", probe.Type)
		}
	}
	return nil
}

func verifyHTTPProbe(ctx context.Context, target string, timeout time.Duration) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("http probe failed: %s returned %s", target, resp.Status)
	}
	return nil
}

func closeBody(c io.Closer) {
	_ = c.Close()
}

func probeTimeout(raw int) time.Duration {
	if raw <= 0 {
		return 5 * time.Second
	}
	return time.Duration(raw) * time.Second
}

func isServiceHealthy(raw string) bool {
	state := strings.ToLower(strings.TrimSpace(raw))
	if state == "" {
		return true
	}
	if state == "running" || strings.Contains(state, "up") || strings.Contains(state, "healthy") {
		return true
	}
	return false
}

func (s *Service) ListRuns(ctx context.Context, limit int, hostAlias, projectRef string) ([]RunRecord, error) {
	return s.opsRepo.ListRuns(ctx, limit, hostAlias, projectRef)
}

func (s *Service) GetRun(ctx context.Context, runID string) (RunRecord, error) {
	return s.opsRepo.GetRun(ctx, runID)
}

func (s *Service) ensureProjectExists(ctx context.Context, hostAlias, projectRef string) error {
	if _, err := s.hostService.Inspect(ctx, hostAlias); err != nil {
		return err
	}
	_, err := s.projectRepo.GetProject(ctx, hostAlias, projectRef)
	return err
}

func mergeProbes(policy []HealthProbe, httpFlags, tcpFlags, cmdFlags []string) []HealthProbe {
	out := make([]HealthProbe, 0, len(policy)+len(httpFlags)+len(tcpFlags)+len(cmdFlags))
	out = append(out, policy...)
	for _, item := range httpFlags {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, HealthProbe{Type: "http", Target: item, Timeout: 5})
		}
	}
	for _, item := range tcpFlags {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, HealthProbe{Type: "tcp", Target: item, Timeout: 5})
		}
	}
	for _, item := range cmdFlags {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, HealthProbe{Type: "cmd", Target: item, Timeout: 5})
		}
	}
	return out
}

func outlierWaitTimeout() time.Duration {
	return 2 * time.Minute
}
