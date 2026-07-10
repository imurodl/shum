package discovery

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/imurodl/shum/internal/projects"
)

type SummaryOptions struct {
	HostAlias string
	Projects  []RuntimeProject
}

func RenderDiscoverSummary(out io.Writer, opts SummaryOptions) {
	_, _ = fmt.Fprintf(out, "Host: %s\n", opts.HostAlias)
	_, _ = fmt.Fprintf(out, "Projects discovered: %d\n", len(opts.Projects))
	_, _ = fmt.Fprintln(out, "Ref\tStatus\tSource\tContext")
	for _, project := range opts.Projects {
		context := project.Directory
		if len(project.ComposeFiles) > 0 {
			context = project.ComposeFiles[0]
		}
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", project.Name, project.Status, project.Source, context)
	}
	_, _ = fmt.Fprintln(out, "")
}

func RenderDiscoverJSON(projects []RuntimeProject) string {
	raw, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func RenderCountByStatus(runtimeProjects []RuntimeProject) map[projects.ProjectStatus]int {
	counts := map[projects.ProjectStatus]int{}
	for _, project := range runtimeProjects {
		counts[project.Status]++
	}
	return counts
}
