package cli

import "github.com/spf13/cobra"

func newProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage compose projects on registered hosts",
	}
	cmd.AddCommand(newProjectDiscoverCommand())
	cmd.AddCommand(newProjectInspectCommand())
	cmd.AddCommand(newProjectPreflightCommand())
	cmd.AddCommand(newProjectPlanCommand())
	cmd.AddCommand(newProjectPolicyCommand())
	cmd.AddCommand(newProjectBackupCommand())
	cmd.AddCommand(newProjectUpgradeCommand())
	cmd.AddCommand(newProjectRunCommand())
	return cmd
}
