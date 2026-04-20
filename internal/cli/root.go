package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shum",
		Short: "Safe Compose upgrades on remote SSH hosts, agent-driveable",
		Long: "shum is a CLI that lets humans and AI agents run safe, recoverable " +
			"Docker Compose upgrades on self-hosted Linux hosts. Every command " +
			"speaks --json with stable error codes — see `shum agent-help`.",
	}

	cmd.AddCommand(newHostCommand())
	cmd.AddCommand(newProjectCommand())
	cmd.AddCommand(newAgentHelpCommand(cmd))
	return cmd
}
