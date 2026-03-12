package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/imurodl/shum/internal/config"
	"github.com/imurodl/shum/internal/hosts"
	"github.com/imurodl/shum/internal/ops"
	"github.com/imurodl/shum/internal/projects"
	"github.com/imurodl/shum/internal/remote"
	"github.com/imurodl/shum/internal/store"
)

func newProjectPreflightCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "preflight <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Run preflight validation before upgrades",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			result, err := opsSvc.Preflight(ctx, alias, projectRef)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, result)
			}
			renderPreflight(cmd, result)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectPlanCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "plan <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Show upgrade plan for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			plan, err := opsSvc.Plan(ctx, alias, projectRef, nil)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, plan)
			}
			renderPlan(cmd, plan)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage per-project safety policy",
	}
	cmd.AddCommand(newProjectPolicyShowCommand())
	cmd.AddCommand(newProjectPolicySetCommand())
	return cmd
}

func newProjectPolicyShowCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "show <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Show active policy for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			policy, err := opsSvc.ResolvePolicy(ctx, alias, projectRef)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, policy)
			}
			renderPolicy(cmd, policy)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectPolicySetCommand() *cobra.Command {
	var (
		requireBackup    bool
		backupCommand    string
		restoreCommand   string
		migrationWarning bool
		healthChecks     []string
	)
	cmd := &cobra.Command{
		Use:   "set <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Update safety policy for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			probes, err := parseHealthProbes(healthChecks)
			if err != nil {
				return err
			}
			policy := ops.ProjectPolicy{
				HostAlias:        alias,
				ProjectRef:       projectRef,
				RequireBackup:    requireBackup,
				BackupCommand:    backupCommand,
				RestoreCommand:   restoreCommand,
				HealthChecks:     probes,
				MigrationWarning: migrationWarning,
			}
			for _, w := range validatePolicy(policy) {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", w)
			}
			return opsSvc.SetPolicy(ctx, policy)
		},
	}
	cmd.Flags().BoolVar(&requireBackup, "require-backup", true, "require a pre-upgrade backup for the project")
	cmd.Flags().StringVar(&backupCommand, "backup-command", "", "command used to create a project backup")
	cmd.Flags().StringVar(&restoreCommand, "restore-command", "", "command used to restore a backup")
	cmd.Flags().BoolVar(&migrationWarning, "migration-warning", false, "require --force before running upgrade")
	cmd.Flags().StringSliceVar(&healthChecks, "health-check", nil, "add health checks in <type>:<target> form (type=http|tcp|cmd)")
	return cmd
}

func newProjectBackupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup operations for a project",
	}
	cmd.AddCommand(newProjectBackupTakeCommand())
	cmd.AddCommand(newProjectBackupListCommand())
	cmd.AddCommand(newProjectBackupRestoreCommand())
	return cmd
}

func newProjectBackupTakeCommand() *cobra.Command {
	var outputJSON bool
	var command string
	cmd := &cobra.Command{
		Use:   "take <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Record a project backup artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			backup, err := opsSvc.TakeBackup(ctx, alias, projectRef, command)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, backup)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Backup %s\n", backup.ArtifactPath)
			fmt.Fprintf(cmd.OutOrStdout(), "SHA: %s\n", backup.ArtifactSHA)
			return nil
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "override policy backup command")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectBackupListCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "list <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "List backup artifacts for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			items, err := opsSvc.ListBackups(ctx, alias, projectRef)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, items)
			}
			for _, item := range items {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"%d\t%s\t%s\t%s\n",
					item.ID,
					item.CreatedAt.Format(time.RFC3339),
					item.ArtifactSHA[:12],
					item.ArtifactPath,
				)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectBackupRestoreCommand() *cobra.Command {
	var command string
	cmd := &cobra.Command{
		Use:   "restore <alias> <project-ref> <artifact-path>",
		Args:  cobra.ExactArgs(3),
		Short: "Restore a backup artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			artifactPath := args[2]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			return opsSvc.RestoreBackup(ctx, alias, projectRef, artifactPath, command)
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "override policy restore command")
	return cmd
}

func newProjectUpgradeCommand() *cobra.Command {
	var (
		force      bool
		skipBackup bool
		dryRun     bool
		outputJSON bool
		httpProbe  []string
		tcpProbe   []string
		cmdProbe   []string
	)
	cmd := &cobra.Command{
		Use:   "upgrade <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Run a safe upgrade for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			opts := ops.UpgradeOptions{
				Force:      force,
				SkipBackup: skipBackup,
				DryRun:     dryRun,
				HttpProbes: httpProbe,
				TcpProbes:  tcpProbe,
				CmdProbes:  cmdProbe,
			}
			result, err := opsSvc.RunUpgrade(ctx, alias, projectRef, opts)
			if err != nil {
				if outputJSON {
					_ = encodeJSON(cmd, result)
				}
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, result)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Run: %s\nStatus: %s\nSummary: %s\n", result.RunID, result.Status, result.Summary)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "override migration-warning blocks")
	cmd.Flags().BoolVar(&skipBackup, "skip-backup", false, "do not create backup during this upgrade")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "plan and validate without mutating")
	cmd.Flags().StringSliceVar(&httpProbe, "http-probe", nil, "add an HTTP probe target")
	cmd.Flags().StringSliceVar(&tcpProbe, "tcp-probe", nil, "add a TCP probe target")
	cmd.Flags().StringSliceVar(&cmdProbe, "cmd-probe", nil, "add a command probe target")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Inspect upgrade run history",
	}
	cmd.AddCommand(newProjectRunListCommand())
	cmd.AddCommand(newProjectRunShowCommand())
	return cmd
}

func newProjectRunListCommand() *cobra.Command {
	var limit int
	var hostFilter string
	var projectFilter string
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List upgrade runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			runs, err := opsSvc.ListRuns(ctx, limit, hostFilter, projectFilter)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, runs)
			}
			for _, run := range runs {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"%s\t%s\t%s\t%s\t%s\n",
					run.RunID,
					run.HostAlias,
					run.ProjectRef,
					run.Status,
					run.StartedAt.Format(time.RFC3339),
				)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "number of runs to show")
	cmd.Flags().StringVar(&hostFilter, "host", "", "filter by host alias")
	cmd.Flags().StringVar(&projectFilter, "project", "", "filter by project reference")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newProjectRunShowCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "show <run-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Show upgrade run details",
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			opsSvc, err := newOpsService()
			if err != nil {
				return err
			}
			run, err := opsSvc.GetRun(ctx, runID)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, run)
			}
			renderRun(cmd, run)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newOpsService() (*ops.Service, error) {
	cfg, err := config.ResolvePaths()
	if err != nil {
		return nil, err
	}
	db, err := store.New(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	_runner := remote.NewRunner(120 * time.Second)
	hostRepo := hosts.NewRepository(db)
	hostSvc := hosts.NewService(hostRepo, _runner)
	projectRepo := projects.NewProjectRepository(db)
	opsRepo := ops.NewRepository(db)
	return ops.NewService(hostSvc, projectRepo, _runner, opsRepo, cfg.ArtifactDir), nil
}

func parseHealthProbes(values []string) ([]ops.HealthProbe, error) {
	out := make([]ops.HealthProbe, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
			out = append(out, ops.HealthProbe{Type: "http", Target: raw, Timeout: 5})
			continue
		}
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid health check format: %s", raw)
		}
		kind := strings.TrimSpace(strings.ToLower(parts[0]))
		target := strings.TrimSpace(parts[1])
		if target == "" {
			return nil, fmt.Errorf("missing target in health check: %s", raw)
		}
		switch kind {
		case "http", "tcp", "cmd":
			out = append(out, ops.HealthProbe{Type: kind, Target: target, Timeout: 5})
		default:
			return nil, fmt.Errorf("unsupported health check type %s", kind)
		}
	}
	return out, nil
}

func renderPreflight(cmd *cobra.Command, result ops.PreflightResult) {
	fmt.Fprintf(cmd.OutOrStdout(), "Host: %s\n", result.HostAlias)
	fmt.Fprintf(cmd.OutOrStdout(), "Passed: %v\n", result.Passed)
	fmt.Fprintf(cmd.OutOrStdout(), "Docker: %s\n", result.DockerVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "Compose: %s\n", result.ComposeVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "Disk available: %d bytes (%s)\n", result.DiskBytesAvail, result.DiskPath)
	for name, status := range result.Checks {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", name, status)
	}
}

func renderPlan(cmd *cobra.Command, plan ops.Plan) {
	fmt.Fprintf(cmd.OutOrStdout(), "Host: %s\n", plan.HostAlias)
	fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", plan.ProjectRef)
	fmt.Fprintf(cmd.OutOrStdout(), "Preflight passed: %v\n", plan.Preflight.Passed)
	for _, item := range plan.Services {
		fmt.Fprintf(
			cmd.OutOrStdout(),
			"%s\timage=%s\tcurrent=%s\ttarget=%s\n",
			item.ServiceName,
			item.Image,
			item.CurrentDigest,
			item.TargetDigest,
		)
	}
	if len(plan.Warnings) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Fprintf(cmd.OutOrStdout(), " - %s\n", warning)
		}
	}
	if len(plan.Blocks) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Blocks:")
		for _, block := range plan.Blocks {
			fmt.Fprintf(cmd.OutOrStdout(), " - %s\n", block)
		}
	}
}

func renderPolicy(cmd *cobra.Command, policy ops.ProjectPolicy) {
	fmt.Fprintf(cmd.OutOrStdout(), "Host: %s\n", policy.HostAlias)
	fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", policy.ProjectRef)
	fmt.Fprintf(cmd.OutOrStdout(), "Require backup: %v\n", policy.RequireBackup)
	fmt.Fprintf(cmd.OutOrStdout(), "Backup command: %s\n", policy.BackupCommand)
	fmt.Fprintf(cmd.OutOrStdout(), "Restore command: %s\n", policy.RestoreCommand)
	fmt.Fprintf(cmd.OutOrStdout(), "Migration warning: %v\n", policy.MigrationWarning)
	for _, probe := range policy.HealthChecks {
		fmt.Fprintf(cmd.OutOrStdout(), "Probe: %s:%s\n", probe.Type, probe.Target)
	}
}

func renderRun(cmd *cobra.Command, run ops.RunRecord) {
	fmt.Fprintf(cmd.OutOrStdout(), "Run: %s\n", run.RunID)
	fmt.Fprintf(cmd.OutOrStdout(), "Host: %s\n", run.HostAlias)
	fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", run.ProjectRef)
	fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", run.Status)
	fmt.Fprintf(cmd.OutOrStdout(), "Started: %s\n", run.StartedAt.Format(time.RFC3339))
	if !run.FinishedAt.IsZero() {
		fmt.Fprintf(cmd.OutOrStdout(), "Finished: %s\n", run.FinishedAt.Format(time.RFC3339))
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Summary: %s\n", run.Summary)
	if run.FailureReason != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Failure: %s\n", run.FailureReason)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Backup artifact: %s\n", run.BackupArtifact)
	fmt.Fprintf(cmd.OutOrStdout(), "Events: %d\n", len(run.Events))
}

func validatePolicy(p ops.ProjectPolicy) []string {
	var warnings []string
	if p.RequireBackup && strings.TrimSpace(p.BackupCommand) == "" {
		warnings = append(warnings, "require-backup is enabled but no --backup-command is set; upgrades will fail at the backup step")
	}
	if strings.TrimSpace(p.BackupCommand) != "" && strings.TrimSpace(p.RestoreCommand) == "" {
		warnings = append(warnings, "backup-command is set but no --restore-command; rollbacks will use docker compose down/up instead of restoring the backup")
	}
	return warnings
}
