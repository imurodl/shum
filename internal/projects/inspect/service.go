package inspect

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/imurodl/shum/internal/hosts"
	"github.com/imurodl/shum/internal/projects"
	"github.com/imurodl/shum/internal/remote"
)

type Service struct {
	runner       *remote.Runner
	hostService  *hosts.Service
	projectRepo  *projects.ProjectRepository
	artifactBase string
}

func NewService(runner *remote.Runner, hostService *hosts.Service, projectRepo *projects.ProjectRepository, artifactBase string) *Service {
	return &Service{
		runner:       runner,
		hostService:  hostService,
		projectRepo:  projectRepo,
		artifactBase: artifactBase,
	}
}

func (s *Service) Inspect(ctx context.Context, hostAlias string, options InspectOptions) (InspectResult, error) {
	host, err := s.hostService.Inspect(ctx, hostAlias)
	if err != nil {
		return InspectResult{}, err
	}

	project, err := s.projectRepo.GetProject(ctx, hostAlias, options.ProjectRef)
	if err != nil {
		return InspectResult{}, err
	}

	projectDir := options.ProjectDir
	if projectDir == "" {
		projectDir = project.ProjectDirectory
	}

	projectName := options.ProjectName
	if projectName == "" {
		projectName = project.ProjectName
	}

	files := options.Files
	if len(files) == 0 && len(project.ComposeFiles) > 0 {
		files = append([]string(nil), project.ComposeFiles...)
	}

	args := []string{}
	if projectName != "" {
		args = append(args, "-p", projectName)
	}
	for _, file := range files {
		args = append(args, "-f", file)
	}
	for _, profile := range options.Profiles {
		args = append(args, "--profile", profile)
	}
	for _, envFile := range options.EnvFiles {
		args = append(args, "--env-file", envFile)
	}

	composePrefix := "docker compose"
	if projectDir != "" {
		composePrefix = fmt.Sprintf("cd %q && %s", projectDir, composePrefix)
	}

	composeCommand := composePrefix
	if len(args) > 0 {
		composeCommand = fmt.Sprintf("%s %s", composePrefix, strings.TrimSpace(strings.Join(args, " ")))
	}
	configCmd := composeCommand + " config --format json"
	configRaw, err := s.runner.Command(host.Alias, configCmd)
	if err != nil {
		project.Status = projects.StatusBlocked
		project.Canonical = false
		_ = s.projectRepo.UpsertProject(ctx, project)
		return InspectResult{
			HostAlias:        host.Alias,
			TrustFingerprint: host.HostKeyFingerprint,
			Project:          project,
			Status:           string(projects.StatusBlocked),
			Reasons:          []string{err.Error()},
		}, nil
	}

	servicesRaw, _ := s.runner.Command(host.Alias, composeCommand+" config --services")
	envRaw, _ := s.runner.Command(host.Alias, composeCommand+" config --environment")
	profilesRaw, _ := s.runner.Command(host.Alias, composeCommand+" config --profiles")
	volumesRaw, _ := s.runner.Command(host.Alias, composeCommand+" config --volumes")
	networksRaw, _ := s.runner.Command(host.Alias, composeCommand+" config --networks")
	psRaw, _ := s.runner.Command(host.Alias, composeCommand+" ps --format json")
	declaredProfiles, activeProfiles := resolveProfileOutputs(profilesRaw, options.Profiles)

	result := InspectResult{
		HostAlias:       host.Alias,
		TrustFingerprint: host.HostKeyFingerprint,
		Project:         project,
		Services:        splitLines(servicesRaw),
		Volumes:         splitLines(volumesRaw),
		Networks:        splitLines(networksRaw),
		Profiles:        declaredProfiles,
		ActiveProfiles:  activeProfiles,
		Status:          string(projects.StatusCanonical),
		Config:          maybeRedactConfig(configRaw, options.ShowConfig),
		Reasons:         []string{},
	}
	if len(options.Profiles) == 0 && containsPotentialProfiles(result.Profiles) {
		result.Status = string(projects.StatusAmbiguous)
		result.Reasons = append(result.Reasons, "profiles require explicit selection")
	}

	if err := s.saveArtifacts(ctx, hostAlias, options.ProjectRef, configRaw, psRaw, ""); err == nil {
		result.Artifact = InspectArtifact{
			ContextJSONPath: filepath.Join(hostAlias, options.ProjectRef, "config.json"),
			RuntimeStatePath: filepath.Join(hostAlias, options.ProjectRef, "runtime.json"),
		}
	}

	project.Status = projects.ProjectStatus(result.Status)
	project.Canonical = result.Status == string(projects.StatusCanonical)
	project.EnvFingerprint = envFingerprint(envRaw)
	_ = s.projectRepo.UpsertProject(ctx, project)
	if options.ShowMounts {
		mounts, _ := s.runner.Command(host.Alias, "docker inspect --type=container --format '{{json .Mounts}}' $(docker ps -q)")
		result.Mounts = splitLines(mounts)
	}
	return result, nil
}

func containsPotentialProfiles(raw []string) bool {
	for _, item := range raw {
		if strings.TrimSpace(item) != "" {
			return true
		}
	}
	return false
}

func envFingerprint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "no-env-metadata"
	}
	sum := 0
	for _, r := range raw {
		sum += int(r)
	}
	return fmt.Sprintf("env-sum:%d", sum)
}

func splitLines(raw string) []string {
	lines := strings.Split(raw, "\n")
	out := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func resolveProfileOutputs(raw string, selected []string) ([]string, []string) {
	declared := splitLines(raw)
	active := append([]string(nil), selected...)
	return declared, active
}

func maybeRedactConfig(raw string, show bool) string {
	if !show {
		return "[hidden]"
	}
	return raw
}
