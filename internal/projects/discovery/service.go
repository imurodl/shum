package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/imurodl/shum/internal/hosts"
	"github.com/imurodl/shum/internal/projects"
	"github.com/imurodl/shum/internal/remote"
)

type Service struct {
	runner      *remote.Runner
	hostService *hosts.Service
	repo        *projects.ProjectRepository
	resolver    *Resolver
}

func NewService(runner *remote.Runner, hostService *hosts.Service, repo *projects.ProjectRepository) *Service {
	return &Service{
		runner:      runner,
		hostService: hostService,
		repo:        repo,
		resolver:    NewResolver(),
	}
}

type DiscoverOptions struct {
	HostAlias string
	Paths     []string
}

func (s *Service) Discover(ctx context.Context, opts DiscoverOptions) ([]RuntimeProject, error) {
	if _, err := s.hostService.Inspect(ctx, opts.HostAlias); err != nil {
		return nil, err
	}

	runtimeProjects, err := s.discoverRuntime(ctx, opts.HostAlias)
	if err == nil {
		for _, item := range runtimeProjects {
			item.Status = classifyFromRuntime(item.Name)
			_ = s.repo.UpsertProject(ctx, projectRecordFromRuntime(opts.HostAlias, item))
		}
		return runtimeProjects, nil
	}

	if len(opts.Paths) > 0 {
		pathProjects, err := s.resolver.Resolve(ctx, ResolveOptions{
			HostAlias: opts.HostAlias,
			Paths:     opts.Paths,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range pathProjects {
			_ = s.repo.UpsertProject(ctx, projectRecordFromRuntime(opts.HostAlias, item))
		}
		return pathProjects, nil
	}

	return nil, fmt.Errorf("discovery failed: %w", err)
}

func (s *Service) discoverRuntime(ctx context.Context, hostAlias string) ([]RuntimeProject, error) {
	raw, err := s.runner.Command(
		hostAlias,
		"docker compose ls --all --format json",
	)
	if err != nil {
		alt, altErr := s.discoverFromContainers(ctx, hostAlias)
		if altErr != nil {
			return nil, err
		}
		return alt, nil
	}

	parsed := parseComposeLSOutput(raw)
	if len(parsed) == 0 {
		return nil, fmt.Errorf("no projects found")
	}
	return parsed, nil
}

func parseComposeLSOutput(raw string) []RuntimeProject {
	type composeLSRow struct {
		Name  string `json:"Name"`
		Files string `json:"ConfigFiles"`
	}

	makeProject := func(row composeLSRow) RuntimeProject {
		project := RuntimeProject{
			Name:       strings.TrimSpace(row.Name),
			Status:     projects.StatusRuntimeOnly,
			Source:     "compose ls",
			RawCommand: "docker compose ls --all --format json",
			Profiles:   []string{},
		}
		if row.Files != "" {
			files := strings.Split(row.Files, ",")
			for i := range files {
				files[i] = strings.TrimSpace(files[i])
			}
			project.ComposeFiles = files
			if len(files) > 0 && files[0] != "" {
				project.Directory = filepath.Dir(files[0])
			}
		}
		return project
	}

	var rows []composeLSRow
	if err := json.Unmarshal([]byte(raw), &rows); err == nil {
		out := make([]RuntimeProject, 0, len(rows))
		for _, row := range rows {
			if strings.TrimSpace(row.Name) == "" {
				continue
			}
			out = append(out, makeProject(row))
		}
		return out
	}

	out := []RuntimeProject{}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row composeLSRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		if strings.TrimSpace(row.Name) == "" {
			continue
		}
		out = append(out, makeProject(row))
	}
	return out
}

func (s *Service) discoverFromContainers(ctx context.Context, hostAlias string) ([]RuntimeProject, error) {
	raw, err := s.runner.Command(
		hostAlias,
		"docker container ls --all --filter label=com.docker.compose.project --format json",
	)
	if err != nil {
		return nil, err
	}

	type containerEntry struct {
		Name  string `json:"Names"`
		Labels map[string]string `json:"Labels"`
	}
	lines := strings.Split(raw, "\n")
	seen := map[string]struct{}{}
	out := make([]RuntimeProject, 0, 16)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry containerEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		project := strings.TrimSpace(entry.Labels["com.docker.compose.project"])
		if project == "" {
			continue
		}
		if _, ok := seen[project]; ok {
			continue
		}
		seen[project] = struct{}{}
		out = append(out, RuntimeProject{
			Name:       project,
			Status:     projects.StatusRuntimeOnly,
			Source:     "container labels",
			RawCommand: "docker container ls --all --filter label=com.docker.compose.project --format json",
			Reason:     "runtime-only discovery from labels",
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("runtime fallback found nothing")
	}
	return out, nil
}

func classifyFromRuntime(projectName string) projects.ProjectStatus {
	_ = projectName
	return projects.StatusRuntimeOnly
}

func projectRecordFromRuntime(hostAlias string, item RuntimeProject) projects.ProjectRecord {
	return projects.ProjectRecord{
		HostAlias:        hostAlias,
		ProjectRef:       item.Name,
		Status:           item.Status,
		Canonical:        false,
		ProjectName:      item.Name,
		ProjectDirectory: item.Directory,
		ComposeFiles:     item.ComposeFiles,
		ActiveProfiles:   item.Profiles,
		EnvFingerprint:   "",
	}
}

func StatusFromComposeState(code int) projects.ProjectStatus {
	if code == 0 {
		return projects.StatusCanonical
	}
	return projects.StatusRuntimeOnly
}
