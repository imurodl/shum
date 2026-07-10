package inspect

import (
	"encoding/json"
	"fmt"
	"io"
)

func RenderSummary(out io.Writer, result InspectResult) {
	_, _ = fmt.Fprintln(out, "Host:", result.HostAlias)
	_, _ = fmt.Fprintln(out, "Project:", result.Project.ProjectRef)
	_, _ = fmt.Fprintln(out, "Canonical status:", result.Status)
	_, _ = fmt.Fprintln(out, "Identity:", result.Project.ProjectName, "dir:", result.Project.ProjectDirectory)
	_, _ = fmt.Fprintln(out, "Services:", len(result.Services))
	_, _ = fmt.Fprintln(out, "Volumes:", len(result.Volumes))
	if len(result.Reasons) > 0 {
		_, _ = fmt.Fprintln(out, "Issues:")
		for _, reason := range result.Reasons {
			_, _ = fmt.Fprintln(out, " -", reason)
		}
	}
}

func RenderJSON(out io.Writer, payload InspectResult) error {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", raw)
	return err
}
