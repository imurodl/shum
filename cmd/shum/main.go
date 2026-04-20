package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/imurodl/shum/internal/cli"
	"github.com/imurodl/shum/internal/shumerr"
)

func main() {
	cmd := cli.NewRootCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	executed, err := cmd.ExecuteC()
	if err == nil {
		os.Exit(0)
	}

	se := shumerr.Classify(err)

	if jsonRequested(executed) {
		_ = cli.EmitError(executed, se)
	} else if se.Hint != "" {
		fmt.Fprintf(os.Stderr, "Error [%s]: %s\nHint: %s\n", se.Code, se.Message, se.Hint)
	} else {
		fmt.Fprintf(os.Stderr, "Error [%s]: %s\n", se.Code, se.Message)
	}

	os.Exit(shumerr.ExitCode(se.Code))
}

func jsonRequested(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	v, err := cmd.Flags().GetBool("json")
	if err != nil {
		return false
	}
	return v
}
