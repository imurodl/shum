package inspect

import "github.com/imurodl/shum/internal/projects"

type InspectOptions struct {
	ProjectRef      string
	ProjectDir      string
	ProjectName     string
	Files           []string
	Profiles        []string
	EnvFiles        []string
	ShowConfig      bool
	ShowMounts      bool
}

type InspectArtifact struct {
	ContextJSONPath  string `json:"context_json"`
	RuntimeStatePath string `json:"runtime_state"`
	MountsPath      string `json:"mounts"`
}

type InspectResult struct {
	HostAlias      string `json:"host"`
	TrustFingerprint string `json:"trust_fingerprint"`
	Project        projects.ProjectRecord `json:"project"`
	Services       []string `json:"services"`
	Volumes        []string `json:"volumes"`
	Networks       []string `json:"networks"`
	Profiles       []string `json:"profiles_declared"`
	ActiveProfiles []string `json:"profiles_active"`
	Status         string   `json:"status"`
	Reasons        []string `json:"reasons"`
	Config         string   `json:"config"`
	Mounts         []string `json:"mounts"`
	Artifact       InspectArtifact `json:"artifact"`
}
