package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/imurodl/shum/internal/shumerr"
)

// Version is set at build time via -ldflags.
var Version = "dev"

const agentContractRevision = "1"

type agentHelpDoc struct {
	Tool         agentTool                `json:"tool"`
	Description  string                   `json:"description"`
	Contract     agentContract            `json:"contract"`
	Commands     []agentCommand           `json:"commands"`
	Errors       map[string]agentError    `json:"errors"`
	OutputShapes map[string]string        `json:"output_shapes"`
	Examples     []agentExample           `json:"examples"`
}

type agentTool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
}

type agentContract struct {
	Revision     string `json:"revision"`
	JSONFlag     string `json:"json_flag"`
	ErrorChannel string `json:"error_channel"`
	ErrorShape   string `json:"error_shape"`
}

type agentCommand struct {
	Path    string       `json:"path"`
	Short   string       `json:"short"`
	Long    string       `json:"long,omitempty"`
	Args    string       `json:"args,omitempty"`
	Flags   []agentFlag  `json:"flags,omitempty"`
	Returns string       `json:"returns,omitempty"`
}

type agentFlag struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

type agentError struct {
	Description string `json:"description"`
	ExitCode    int    `json:"exit_code"`
}

type agentExample struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

func newAgentHelpCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent-help",
		Short: "Emit the full CLI surface as JSON for AI agents",
		Long: "Returns a single JSON document describing every command, every flag, " +
			"every error code, and the JSON output shape per command. " +
			"Designed to be loaded once into an agent's context.",
		RunE: func(cmd *cobra.Command, args []string) error {
			doc := buildAgentHelp(root)
			return emitJSON(cmd, doc)
		},
	}
	return cmd
}

func buildAgentHelp(root *cobra.Command) agentHelpDoc {
	return agentHelpDoc{
		Tool: agentTool{
			Name:    "shum",
			Version: Version,
			Repo:    "https://github.com/imurodl/shum",
		},
		Description: "Safe Docker Compose upgrades on remote SSH hosts. " +
			"All commands accept --json for structured output. " +
			"On failure, an envelope `{\"error\":{\"code\":...}}` is written to stderr and the process exits non-zero.",
		Contract: agentContract{
			Revision:     agentContractRevision,
			JSONFlag:     "--json",
			ErrorChannel: "stderr",
			ErrorShape:   `{"error":{"code":"<stable_code>","message":"...","hint":"...","details":{...}}}`,
		},
		Commands:     collectCommands(root, ""),
		Errors:       collectErrors(),
		OutputShapes: outputShapes(),
		Examples:     examples(),
	}
}

func collectCommands(cmd *cobra.Command, prefix string) []agentCommand {
	var out []agentCommand
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
			continue
		}
		path := strings.TrimSpace(prefix + " " + child.Name())
		if !child.HasSubCommands() || child.Runnable() {
			out = append(out, agentCommand{
				Path:    path,
				Short:   child.Short,
				Long:    cleanLong(child.Long),
				Args:    child.Use,
				Flags:   collectFlags(child),
				Returns: outputShapes()[path],
			})
		}
		if child.HasSubCommands() {
			out = append(out, collectCommands(child, path)...)
		}
	}
	return out
}

func collectFlags(cmd *cobra.Command) []agentFlag {
	var out []agentFlag
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		out = append(out, agentFlag{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		})
	})
	return out
}

func cleanLong(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}

func collectErrors() map[string]agentError {
	out := make(map[string]agentError, len(shumerr.AllCodes()))
	for _, code := range shumerr.AllCodes() {
		out[code] = agentError{
			Description: shumerr.Description(code),
			ExitCode:    shumerr.ExitCode(code),
		}
	}
	return out
}

// outputShapes returns a one-line description of the JSON shape each
// command emits on success, keyed by full command path. Hand-curated so
// agents have a stable target type to parse against.
func outputShapes() map[string]string {
	return map[string]string{
		"host register":         "ops.HostOutput {alias, hostname, user, port, last_verified_at, remote_os, remote_arch, docker_version, compose_version, known_hosts_files, fingerprint}",
		"host list":             "[]hosts.Host",
		"host inspect":          "hosts.Host",
		"project discover":      "[]discovery.RuntimeProject {name, status, directory, compose_files, profiles}",
		"project inspect":       "inspect.Result {project_ref, project_name, project_directory, compose_files, profiles, env_files, config?, mounts?}",
		"project preflight":     "ops.PreflightResult {host_alias, docker_available, compose_available, docker_version, compose_version, disk_bytes_available, disk_path, permissions_ok, checks, passed}",
		"project plan":          "ops.Plan {host_alias, project_ref, run_id, preflight, services, actions, policy, created_at, warnings, blocks}",
		"project policy show":   "ops.ProjectPolicy {require_backup, backup_command, restore_command, health_checks, migration_warning}",
		"project policy set":    "(no body; non-zero exit on failure)",
		"project backup take":   "ops.BackupResult {id, host_alias, project_ref, artifact_path, artifact_sha256, command, created_at, size_bytes}",
		"project backup list":   "[]ops.BackupResult",
		"project backup restore": "(no body; non-zero exit on failure)",
		"project upgrade":       "ops.UpgradeResult {run_id, status: planned|running|success|failed|rolled_back, summary}",
		"project run list":      "[]ops.RunRecord",
		"project run show":      "ops.RunRecord {id, run_id, host_alias, project_ref, status, started_at, finished_at, preflight, plan, summary, backup_artifact, failure_reason, events: []RunEvent}",
		"agent-help":            "shum.agentHelpDoc (this document)",
	}
}

func examples() []agentExample {
	return []agentExample{
		{Title: "Discover what to operate on", Command: "shum host list --json"},
		{Title: "Plan an upgrade and read the JSON before acting", Command: "shum project plan prod web --json"},
		{Title: "Dry-run an upgrade", Command: "shum project upgrade prod web --dry-run --json"},
		{Title: "Real upgrade with backup enforced by policy", Command: "shum project upgrade prod web --json"},
		{Title: "Inspect the most recent run", Command: "shum project run list --limit 1 --json"},
	}
}
