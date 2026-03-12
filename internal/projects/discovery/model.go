package discovery

import "github.com/imurodl/shum/internal/projects"

type RuntimeProject struct {
	Name       string            `json:"name"`
	Status     projects.ProjectStatus `json:"status"`
	Source     string            `json:"source"`
	Services   int               `json:"services"`
	Directory  string            `json:"project_directory"`
	ComposeFiles []string        `json:"compose_files"`
	Profiles   []string          `json:"profiles"`
	RawCommand string            `json:"raw_command"`
	Reason     string            `json:"reason"`
}
