package hosts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/imurodl/shum/internal/remote"
	ssh "github.com/imurodl/shum/internal/ssh"
)

type Service struct {
	repo   *Repository
	runner *remote.Runner
}

func NewService(repo *Repository, runner *remote.Runner) *Service {
	return &Service{
		repo:   repo,
		runner: runner,
	}
}

func (s *Service) Register(ctx context.Context, alias string) (Host, error) {
	resolved, err := ssh.ParseResolvedAlias(alias)
	if err != nil {
		return Host{}, fmt.Errorf("alias resolution failed: %w", err)
	}

	files := resolved.KnownHostFiles
	if len(files) == 0 {
		return Host{}, fmt.Errorf("no known_hosts file configured for ssh alias %s", alias)
	}

	fingerprint, err := ssh.VerifyHostKey(resolved.Hostname, resolved.Port, files)
	if err != nil {
		return Host{}, err
	}

	probe, err := ssh.ProbeAlias(alias, s.runner)
	if err != nil {
		return Host{}, fmt.Errorf("host probe failed: %w", err)
	}
	if !strings.EqualFold(probe.OS, "linux") {
		return Host{}, fmt.Errorf("target is not Linux: %s", probe.OS)
	}

	host := Host{
		Alias:             alias,
		Hostname:          resolved.Hostname,
		UserName:          resolved.User,
		Port:              resolved.Port,
		KnownHostsFiles:   files,
		HostKeyFingerprint: fingerprint,
		RemoteOS:          probe.OS,
		RemoteArch:        probe.Arch,
		DockerVersion:     probe.DockerVersion,
		ComposeVersion:    probe.ComposeVersion,
		LastVerifiedAt:     time.Now().UTC(),
	}

	if err := s.repo.Upsert(ctx, host); err != nil {
		return Host{}, err
	}
	return host, nil
}

func (s *Service) List(ctx context.Context) ([]Host, error) {
	return s.repo.List(ctx)
}

func (s *Service) Inspect(ctx context.Context, alias string) (Host, error) {
	return s.repo.Get(ctx, alias)
}
