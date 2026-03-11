package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/your-org/shum/internal/projects"
)

type SummaryOptions struct {
	HostAlias string
	Projects  []RuntimeProject
}

func RenderDiscoverSummary(out io.Writer, opts SummaryOptions) {
	fmt.Fprintf(out, "Host: %s\n", opts.HostAlias)
	fmt.Fprintf(out, "Projects discovered: %d\n", len(opts.Projects))
	fmt.Fprintln(out, "Ref\tStatus\tSource\tContext")
	for _, project := range opts.Projects {
		context := project.Directory
		if len(project.ComposeFiles) > 0 {
			context = project.ComposeFiles[0]
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", project.Name, project.Status, project.Source, context)
	}
	fmt.Fprintln(out, "")
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

func statusBadge(counts map[projects.ProjectStatus]int) string {
	parts := make([]string, 0, len(counts))
	for key, value := range counts {
		parts = append(parts, string(key)+":"+strconv.Itoa(value))
	}
	return strings.Join(parts, ", ")
}
