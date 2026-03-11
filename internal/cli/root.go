package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shum",
		Short: "Self-Host Upgrade Manager CLI",
		Long:  "shum helps operators run safe, recoverable upgrades across self-hosted Linux hosts.",
	}

	cmd.AddCommand(newHostCommand())
	cmd.AddCommand(newProjectCommand())
	return cmd
}
