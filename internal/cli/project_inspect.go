package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/your-org/shum/internal/config"
	"github.com/your-org/shum/internal/hosts"
	"github.com/your-org/shum/internal/projects"
	"github.com/your-org/shum/internal/projects/discovery"
	"github.com/your-org/shum/internal/projects/inspect"
	"github.com/your-org/shum/internal/remote"
	"github.com/your-org/shum/internal/store"
)

func newProjectInspectCommand() *cobra.Command {
	var (
		outputJSON   bool
		projectDir   string
		projectName  string
		fileFlags    []string
		profileFlags []string
		envFlags     []string
		showConfig   bool
		showMounts   bool
	)

	cmd := &cobra.Command{
		Use:   "inspect <alias> <project-ref>",
		Args:  cobra.ExactArgs(2),
		Short: "Inspect canonical config and risk surfaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			hostAlias := args[0]
			projectRef := args[1]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			inspectSvc, _, err := newInspectService()
			if err != nil {
				return err
			}
			out := inspect.InspectOptions{
				ProjectRef:  projectRef,
				ProjectDir:  projectDir,
				ProjectName: projectName,
				Files:       fileFlags,
				Profiles:    profileFlags,
				EnvFiles:    envFlags,
				ShowConfig:  showConfig,
				ShowMounts:  showMounts,
			}
			result, err := inspectSvc.Inspect(ctx, hostAlias, out)
			if err != nil {
				return err
			}
			if outputJSON {
				return inspect.RenderJSON(cmd.OutOrStdout(), result)
			}
			inspect.RenderSummary(cmd.OutOrStdout(), result)
			if showConfig {
				fmt.Fprintln(cmd.OutOrStdout(), "")
				fmt.Fprintln(cmd.OutOrStdout(), "Config:")
				fmt.Fprintln(cmd.OutOrStdout(), result.Config)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&projectDir, "project-directory", "", "explicit compose project directory")
	cmd.Flags().StringVar(&projectName, "project-name", "", "explicit compose project name")
	cmd.Flags().StringSliceVar(&fileFlags, "file", nil, "ordered compose files")
	cmd.Flags().StringSliceVar(&profileFlags, "profile", nil, "compose profiles to activate")
	cmd.Flags().StringSliceVar(&envFlags, "env-file", nil, "compose env files")
	cmd.Flags().BoolVar(&showConfig, "show-config", false, "show rendered docker compose config")
	cmd.Flags().BoolVar(&showMounts, "show-mounts", false, "show mount surfaces")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")

	return cmd
}

func newInspectService() (*inspect.Service, *discovery.Service, error) {
	cfg, err := config.ResolvePaths()
	if err != nil {
		return nil, nil, err
	}
	db, err := store.New(cfg.DatabasePath)
	if err != nil {
		return nil, nil, err
	}
	runner := remote.NewRunner(60 * time.Second)
	hostRepo := hosts.NewRepository(db)
	hostSvc := hosts.NewService(hostRepo, runner)
	projectRepo := projects.NewProjectRepository(db)
	projectDiscover := discovery.NewService(runner, hostSvc, projectRepo)
	return inspect.NewService(runner, hostSvc, projectRepo, cfg.ArtifactDir), projectDiscover, nil
}
