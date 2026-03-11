package projects

type ProjectStatus string

const (
	StatusCanonical = ProjectStatus("canonical")
	StatusRuntimeOnly = ProjectStatus("runtime_only")
	StatusAmbiguous = ProjectStatus("ambiguous")
	StatusBlocked = ProjectStatus("blocked")
)

