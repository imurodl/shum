package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/imurodl/shum/internal/config"
	"github.com/imurodl/shum/internal/hosts"
	"github.com/imurodl/shum/internal/remote"
	"github.com/imurodl/shum/internal/store"
)

type hostOutput struct {
	Alias        string   `json:"alias"`
	Hostname     string   `json:"hostname"`
	User         string   `json:"user"`
	Port         int      `json:"port"`
	LastVerified string   `json:"last_verified_at"`
	RemoteOS     string   `json:"remote_os"`
	RemoteArch   string   `json:"remote_arch"`
	Docker       string   `json:"docker_version"`
	Compose      string   `json:"compose_version"`
	KnownHosts   []string `json:"known_hosts_files"`
	Fingerprint  string   `json:"fingerprint"`
}

func newHostCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host",
		Short: "Manage SSH hosts",
	}
	cmd.AddCommand(newHostRegisterCommand())
	cmd.AddCommand(newHostListCommand())
	cmd.AddCommand(newHostInspectCommand())
	return cmd
}

func newHostRegisterCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "register <alias>",
		Args:  cobra.ExactArgs(1),
		Short: "Register a host by SSH alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			s, err := newHostService()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			host, err := s.Register(ctx, alias)
			if err != nil {
				return err
			}
			if outputJSON {
				out := hostOutput{
					Alias:        host.Alias,
					Hostname:     host.Hostname,
					User:         host.UserName,
					Port:         host.Port,
					LastVerified: host.LastVerifiedAt.UTC().Format(time.RFC3339),
					RemoteOS:     host.RemoteOS,
					RemoteArch:   host.RemoteArch,
					Docker:       host.DockerVersion,
					Compose:      host.ComposeVersion,
					KnownHosts:   host.KnownHostsFiles,
					Fingerprint:  host.HostKeyFingerprint,
				}
				return encodeJSON(cmd, out)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Registered host %s (%s:%d) as %s\n", host.Alias, host.Hostname, host.Port, host.RemoteOS)
			fmt.Fprintf(cmd.OutOrStdout(), "Docker: %s\n", host.DockerVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "Compose: %s\n", host.ComposeVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "Last verified: %s\n", host.LastVerifiedAt.Format(time.RFC3339))
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newHostListCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newHostService()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			rows, err := s.List(ctx)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, rows)
			}
			for _, row := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s:%d\t%s\n", row.Alias, row.Hostname, row.UserName, row.Port, row.RemoteOS)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newHostInspectCommand() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "inspect <alias>",
		Args:  cobra.ExactArgs(1),
		Short: "Inspect a registered host",
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			s, err := newHostService()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			host, err := s.Inspect(ctx, alias)
			if err != nil {
				return err
			}
			if outputJSON {
				return encodeJSON(cmd, host)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Host: %s\n", host.Alias)
			fmt.Fprintf(cmd.OutOrStdout(), "Resolved: %s:%d\n", host.Hostname, host.Port)
			fmt.Fprintf(cmd.OutOrStdout(), "User: %s\n", host.UserName)
			fmt.Fprintf(cmd.OutOrStdout(), "Trust: %s\n", host.TrustSummary())
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "output machine-readable JSON")
	return cmd
}

func newHostService() (*hosts.Service, error) {
	cfg, err := config.ResolvePaths()
	if err != nil {
		return nil, err
	}
	st, err := store.New(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	repo := hosts.NewRepository(st)
	runner := remote.NewRunner(20 * time.Second)
	return hosts.NewService(repo, runner), nil
}

func encodeJSON(cmd *cobra.Command, payload any) error {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(append(raw, '\n'))
	return err
}
