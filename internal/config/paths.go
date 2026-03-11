package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Directories struct {
	ConfigDir     string
	DataDir       string
	ArtifactDir   string
	DatabasePath  string
	KnownHostsDir string
}

var (
	once    sync.Once
	cached  Directories
	loadErr error
)

func ResolvePaths() (Directories, error) {
	once.Do(func() {
		cached, loadErr = resolvePaths()
	})
	return cached, loadErr
}

func resolvePaths() (Directories, error) {
	const app = "shum"
	configHome, err := os.UserConfigDir()
	if err != nil {
		return Directories{}, err
	}
	dataHome, err := os.UserCacheDir()
	if err != nil {
		return Directories{}, err
	}

	configDir := filepath.Join(configHome, app)
	dataDir := filepath.Join(dataHome, app)
	artifactDir := filepath.Join(dataDir, "artifacts")
	dbPath := filepath.Join(dataDir, "state.db")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return Directories{}, err
	}
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		return Directories{}, err
	}

	knownHostsDir := strings.TrimSpace(os.Getenv("SHUM_KNOWN_HOSTS"))
	if knownHostsDir == "" {
		knownHostsDir = filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	}

	if dataDir == "" || configDir == "" {
		return Directories{}, errors.New("failed to resolve required config paths")
	}

	return Directories{
		ConfigDir:     configDir,
		DataDir:       dataDir,
		ArtifactDir:   artifactDir,
		DatabasePath:  dbPath,
		KnownHostsDir: knownHostsDir,
	}, nil
}

func DatabasePath() (string, error) {
	dirs, err := ResolvePaths()
	if err != nil {
		return "", err
	}
	return dirs.DatabasePath, nil
}
