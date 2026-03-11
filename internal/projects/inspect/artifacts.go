package inspect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func (s *Service) saveArtifacts(ctx context.Context, hostAlias, projectRef string, configRaw, runtimeRaw, mountsRaw string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	base := filepath.Join(s.artifactBase, hostAlias, projectRef)
	if err := os.MkdirAll(base, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "config.json"), []byte(configRaw), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "runtime.json"), []byte(runtimeRaw), 0o644); err != nil {
		return err
	}
	if mountsRaw != "" {
		_ = os.WriteFile(filepath.Join(base, "mounts.json"), []byte(mountsRaw), 0o644)
	}
	_ = shaString(configRaw)
	return nil
}

func shaString(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func ArtifactPath(base, hostAlias, projectRef, name string) string {
	return filepath.Join(base, hostAlias, projectRef, name)
}

func MustArtifact(base, hostAlias, projectRef string) string {
	return fmt.Sprintf("%s/%s/%s", base, hostAlias, projectRef)
}
