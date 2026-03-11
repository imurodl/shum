package cli

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/your-org/shum/internal/config"
	"github.com/your-org/shum/internal/hosts"
	"github.com/your-org/shum/internal/projects"
	"github.com/your-org/shum/internal/projects/discovery"
	"github.com/your-org/shum/internal/remote"
	"github.com/your-org/shum/internal/store"
)

func newProjectDiscoverCommand() *cobra.Command {
	var outputJSON bool
	var paths []string
	cmd := &cobra.Command{
		Use:   "discover <alias>",
		Args:  cobra.ExactArgs(1),
		Short: "Discover compose projects on a registered host",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			discover, repo, _, err := newProjectServices()
			if err != nil {
				return err
			}

			for i := range paths {
				var err error
				paths[i], err = filepath.Abs(paths[i])
				if err != nil {
					return err
				}
			}
			results, err := discover.Discover(ctx, discovery.DiscoverOptions{
				HostAlias: alias,
				Paths:     paths,
			})
			if err != nil {
				return err
			}
			for _, p := range results {
				_ = repo.UpsertProject(ctx, mapToProjectRecord(alias, p))
			}
			if outputJSON {
				raw, err := json.MarshalIndent(results, "", "  ")
				if err != nil {
					return err
				}
				_, _ = cmd.OutOrStdout().Write(raw)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
				return nil
			}
			discovery.RenderDiscoverSummary(cmd.OutOrStdout(), discovery.SummaryOptions{
				HostAlias: alias,
				Projects:  results,
			})
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&paths, "path", nil, "explicit project directories to probe")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func mapToProjectRecord(alias string, source discovery.RuntimeProject) projects.ProjectRecord {
	return projects.ProjectRecord{
		HostAlias: alias,
		ProjectRef: source.Name,
		Status: source.Status,
		Canonical: source.Status == projects.StatusCanonical,
		ProjectName: source.Name,
		ProjectDirectory: source.Directory,
		ComposeFiles: source.ComposeFiles,
		ActiveProfiles: source.Profiles,
	}
}

func newProjectServices() (*discovery.Service, *projects.ProjectRepository, *hosts.Service, error) {
	cfg, err := config.ResolvePaths()
	if err != nil {
		return nil, nil, nil, err
	}
	db, err := store.New(cfg.DatabasePath)
	if err != nil {
		return nil, nil, nil, err
	}
	runner := remote.NewRunner(60 * time.Second)
	hostRepo := hosts.NewRepository(db)
	hostSvc := hosts.NewService(hostRepo, runner)
	projectRepo := projects.NewProjectRepository(db)
	service := discovery.NewService(runner, hostSvc, projectRepo)
	return service, projectRepo, hostSvc, nil
}
