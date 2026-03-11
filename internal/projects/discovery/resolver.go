package discovery

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"github.com/your-org/shum/internal/projects"
)

var defaultComposeFilenames = []string{
	"compose.yaml",
	"compose.yml",
	"docker-compose.yaml",
	"docker-compose.yml",
}

type ResolveOptions struct {
	HostAlias  string
	ProjectRef string
	Paths     []string
}

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(ctx context.Context, opts ResolveOptions) ([]RuntimeProject, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	output := []RuntimeProject{}
	for _, dir := range opts.Paths {
		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("path error: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("path is not a directory: %s", dir)
		}

		found := []string{}
		for _, file := range defaultComposeFilenames {
			full := filepath.Join(dir, file)
			if _, err := os.Stat(full); err == nil {
				found = append(found, full)
			}
		}
		sort.Strings(found)
		projectName := opts.ProjectRef
		if projectName == "" {
			projectName = filepath.Base(dir)
		}

		if len(found) == 0 {
			continue
		}
		if len(found) > 1 {
			output = append(output, RuntimeProject{
				Name:       projectName,
				Status:     projects.StatusAmbiguous,
				Source:     "path",
				Directory:  dir,
				ComposeFiles: found,
				Reason:     "multiple compose files found; use --file to set exact order",
			})
			continue
		}
		output = append(output, RuntimeProject{
			Name:       projectName,
			Status:     projects.StatusRuntimeOnly,
			Source:     "path",
			Directory:  dir,
			ComposeFiles: found,
			Reason:     "path discovered without runtime context",
		})
	}
	if len(output) == 0 {
		return nil, fs.ErrNotExist
	}
	return output, nil
}
