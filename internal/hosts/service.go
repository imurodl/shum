package hosts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/imurodl/shum/internal/remote"
	"github.com/imurodl/shum/internal/shumerr"
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
		return Host{}, shumerr.Wrap(shumerr.CodeSSHConfigInvalid, err, fmt.Sprintf("alias resolution failed for %q", alias)).
			WithHint("check that the alias appears in your ~/.ssh/config").
			WithDetails(map[string]any{"alias": alias})
	}

	files := resolved.KnownHostFiles
	if len(files) == 0 {
		return Host{}, shumerr.Newf(shumerr.CodeKnownHostsMissing, "no known_hosts file configured for ssh alias %s", alias).
			WithHint("add a UserKnownHostsFile entry for this host in ~/.ssh/config").
			WithDetails(map[string]any{"alias": alias})
	}

	fingerprint, err := ssh.VerifyHostKey(resolved.Hostname, resolved.Port, files)
	if err != nil {
		return Host{}, shumerr.Wrap(shumerr.CodeHostUnverified, err, "").
			WithDetails(map[string]any{"alias": alias, "hostname": resolved.Hostname, "port": resolved.Port})
	}

	probe, err := ssh.ProbeAlias(alias, s.runner)
	if err != nil {
		// Pass through if probe already returned a coded error; otherwise wrap.
		if _, ok := shumerr.From(err); ok {
			return Host{}, err
		}
		return Host{}, shumerr.Wrap(shumerr.CodeHostUnreachable, err, fmt.Sprintf("host probe failed for %q", alias)).
			WithDetails(map[string]any{"alias": alias})
	}
	if !strings.EqualFold(probe.OS, "linux") {
		return Host{}, shumerr.Newf(shumerr.CodeHostNotLinux, "target is not Linux: %s", probe.OS).
			WithDetails(map[string]any{"alias": alias, "remote_os": probe.OS})
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
