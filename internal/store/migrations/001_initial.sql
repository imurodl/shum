CREATE TABLE IF NOT EXISTS hosts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alias TEXT NOT NULL UNIQUE,
    hostname TEXT NOT NULL,
    user_name TEXT NOT NULL,
    port INTEGER NOT NULL,
    known_hosts_files TEXT NOT NULL,
    host_key_fingerprint TEXT,
    remote_os TEXT,
    remote_arch TEXT,
    docker_version TEXT,
    compose_version TEXT,
    last_verified_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_alias TEXT NOT NULL,
    project_ref TEXT NOT NULL,
    status TEXT NOT NULL,
    canonical BOOLEAN NOT NULL DEFAULT 0,
    project_name TEXT,
    project_directory TEXT,
    compose_files TEXT,
    active_profiles TEXT,
    env_fingerprint TEXT,
    discovered_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY(host_alias) REFERENCES hosts(alias) ON DELETE CASCADE,
    UNIQUE(host_alias, project_ref)
);

CREATE TABLE IF NOT EXISTS snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    artifact_type TEXT NOT NULL,
    artifact_path TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS discovery_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_alias TEXT NOT NULL,
    raw_payload TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY(host_alias) REFERENCES hosts(alias) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_policies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_alias TEXT NOT NULL,
    project_ref TEXT NOT NULL,
    require_backup INTEGER NOT NULL DEFAULT 1,
    backup_command TEXT NOT NULL DEFAULT '',
    restore_command TEXT NOT NULL DEFAULT '',
    health_checks TEXT NOT NULL DEFAULT '[]',
    migration_warning INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(host_alias, project_ref)
);

CREATE TABLE IF NOT EXISTS upgrade_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL UNIQUE,
    host_alias TEXT NOT NULL,
    project_ref TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TEXT,
    finished_at TEXT,
    preflight JSON,
    plan JSON NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    backup_artifact TEXT,
    failure_reason TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS upgrade_run_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    detail TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY(run_id) REFERENCES upgrade_runs(run_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS backups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_alias TEXT NOT NULL,
    project_ref TEXT NOT NULL,
    artifact_path TEXT NOT NULL,
    artifact_sha256 TEXT NOT NULL,
    command TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
