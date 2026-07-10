package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setHome points every home-derived lookup at a throwaway directory so the test
// never touches the real user config/cache dirs. os.UserConfigDir/UserCacheDir
// derive from HOME on darwin and from the XDG_* vars on linux, so we set all.
func setHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "cache"))
	return home
}

func TestResolvePathsStructure(t *testing.T) {
	home := setHome(t)
	t.Setenv("SHUM_KNOWN_HOSTS", "")

	dirs, err := resolvePaths()
	if err != nil {
		t.Fatalf("resolvePaths: %v", err)
	}

	if filepath.Base(dirs.ConfigDir) != "shum" {
		t.Errorf("ConfigDir = %q, want basename shum", dirs.ConfigDir)
	}
	if filepath.Base(dirs.DataDir) != "shum" {
		t.Errorf("DataDir = %q, want basename shum", dirs.DataDir)
	}
	if want := filepath.Join(dirs.DataDir, "artifacts"); dirs.ArtifactDir != want {
		t.Errorf("ArtifactDir = %q, want %q", dirs.ArtifactDir, want)
	}
	if want := filepath.Join(dirs.DataDir, "state.db"); dirs.DatabasePath != want {
		t.Errorf("DatabasePath = %q, want %q", dirs.DatabasePath, want)
	}
	if !strings.HasPrefix(dirs.KnownHostsDir, home) {
		t.Errorf("KnownHostsDir = %q, want a path under HOME %q", dirs.KnownHostsDir, home)
	}

	// ConfigDir and ArtifactDir must be created as a side effect.
	for _, d := range []string{dirs.ConfigDir, dirs.ArtifactDir} {
		fi, err := os.Stat(d)
		if err != nil || !fi.IsDir() {
			t.Errorf("expected directory %q to be created (err=%v)", d, err)
		}
	}
}

func TestResolvePathsKnownHostsOverride(t *testing.T) {
	setHome(t)
	custom := filepath.Join(t.TempDir(), "custom_known_hosts")
	t.Setenv("SHUM_KNOWN_HOSTS", custom)

	dirs, err := resolvePaths()
	if err != nil {
		t.Fatalf("resolvePaths: %v", err)
	}
	if dirs.KnownHostsDir != custom {
		t.Errorf("KnownHostsDir = %q, want override %q", dirs.KnownHostsDir, custom)
	}
}
