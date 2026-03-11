package inspect

import (
	"encoding/json"
	"fmt"
	"io"
)

func RenderSummary(out io.Writer, result InspectResult) {
	fmt.Fprintln(out, "Host:", result.HostAlias)
	fmt.Fprintln(out, "Project:", result.Project.ProjectRef)
	fmt.Fprintln(out, "Canonical status:", result.Status)
	fmt.Fprintln(out, "Identity:", result.Project.ProjectName, "dir:", result.Project.ProjectDirectory)
	fmt.Fprintln(out, "Services:", len(result.Services))
	fmt.Fprintln(out, "Volumes:", len(result.Volumes))
	if len(result.Reasons) > 0 {
		fmt.Fprintln(out, "Issues:")
		for _, reason := range result.Reasons {
			fmt.Fprintln(out, " -", reason)
		}
	}
	if len(result.Config) > 0 {
		fmt.Fprintln(out, "Config: hidden by default")
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
