package hosts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/imurodl/shum/internal/shumerr"
	"github.com/imurodl/shum/internal/store"
)

type Host struct {
	ID                int64
	Alias             string
	Hostname          string
	UserName          string
	Port              int
	KnownHostsFiles   []string
	HostKeyFingerprint string
	RemoteOS          string
	RemoteArch        string
	DockerVersion     string
	ComposeVersion    string
	LastVerifiedAt    time.Time
}

type Repository struct {
	db *store.Store
}

func NewRepository(db *store.Store) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Upsert(ctx context.Context, host Host) error {
	payload, err := json.Marshal(host.KnownHostsFiles)
	if err != nil {
		return err
	}

	query := `
INSERT INTO hosts (
	alias, hostname, user_name, port, known_hosts_files, host_key_fingerprint,
	remote_os, remote_arch, docker_version, compose_version, last_verified_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(alias) DO UPDATE SET
	hostname=excluded.hostname,
	user_name=excluded.user_name,
	port=excluded.port,
	known_hosts_files=excluded.known_hosts_files,
	host_key_fingerprint=excluded.host_key_fingerprint,
	remote_os=excluded.remote_os,
	remote_arch=excluded.remote_arch,
	docker_version=excluded.docker_version,
	compose_version=excluded.compose_version,
	last_verified_at=excluded.last_verified_at,
	updated_at=strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
`
	_, err = r.db.DB().ExecContext(
		ctx,
		query,
		host.Alias,
		host.Hostname,
		host.UserName,
		host.Port,
		string(payload),
		host.HostKeyFingerprint,
		host.RemoteOS,
		host.RemoteArch,
		host.DockerVersion,
		host.ComposeVersion,
		host.LastVerifiedAt.Format(time.RFC3339),
	)
	return err
}

func (r *Repository) List(ctx context.Context) ([]Host, error) {
	rows, err := r.db.DB().QueryContext(ctx, `SELECT id, alias, hostname, user_name, port, known_hosts_files, host_key_fingerprint, remote_os, remote_arch, docker_version, compose_version, last_verified_at FROM hosts ORDER BY alias`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hosts := make([]Host, 0, 32)
	for rows.Next() {
		var h Host
		var rawFiles string
		var verifiedRaw string
		err := rows.Scan(
			&h.ID,
			&h.Alias,
			&h.Hostname,
			&h.UserName,
			&h.Port,
			&rawFiles,
			&h.HostKeyFingerprint,
			&h.RemoteOS,
			&h.RemoteArch,
			&h.DockerVersion,
			&h.ComposeVersion,
			&verifiedRaw,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(rawFiles), &h.KnownHostsFiles); err != nil {
			return nil, fmt.Errorf("corrupt known_hosts_files for %s: %w", h.Alias, err)
		}
		if parsed, err := time.Parse(time.RFC3339, verifiedRaw); err == nil {
			h.LastVerifiedAt = parsed
		}
		hosts = append(hosts, h)
	}
	return hosts, rows.Err()
}

func (r *Repository) Get(ctx context.Context, alias string) (Host, error) {
	var h Host
	var rawFiles string
	var verifiedRaw string

	row := r.db.DB().QueryRowContext(ctx,
		`SELECT id, alias, hostname, user_name, port, known_hosts_files, host_key_fingerprint, remote_os, remote_arch, docker_version, compose_version, last_verified_at
		 FROM hosts WHERE alias = ?`,
		alias,
	)
	err := row.Scan(
		&h.ID,
		&h.Alias,
		&h.Hostname,
		&h.UserName,
		&h.Port,
		&rawFiles,
		&h.HostKeyFingerprint,
		&h.RemoteOS,
		&h.RemoteArch,
		&h.DockerVersion,
		&h.ComposeVersion,
		&verifiedRaw,
	)
	if err == sql.ErrNoRows {
		return Host{}, shumerr.Newf(shumerr.CodeHostNotFound, "host not found: %s", alias).
			WithHint("register it with `shum host register " + alias + "`").
			WithDetails(map[string]any{"alias": alias})
	}
	if err != nil {
		return Host{}, err
	}
	if err := json.Unmarshal([]byte(rawFiles), &h.KnownHostsFiles); err != nil {
		return Host{}, fmt.Errorf("corrupt known_hosts_files for %s: %w", h.Alias, err)
	}
	if parsed, err := time.Parse(time.RFC3339, verifiedRaw); err == nil {
		h.LastVerifiedAt = parsed
	}
	return h, nil
}

func (h Host) TrustSummary() string {
	trust := fmt.Sprintf("last verified %s", h.LastVerifiedAt.Format(time.RFC3339))
	if strings.TrimSpace(h.HostKeyFingerprint) != "" {
		trust = fmt.Sprintf("%s (fingerprint %s)", trust, h.HostKeyFingerprint)
	}
	return trust
}
