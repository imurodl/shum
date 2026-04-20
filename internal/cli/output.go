package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/imurodl/shum/internal/shumerr"
)

// jsonModeRequested reports whether the caller asked for machine-readable
// JSON on this command (or any ancestor). It is safe to call on commands
// that do not declare a --json flag.
func jsonModeRequested(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	if v, err := cmd.Flags().GetBool("json"); err == nil {
		return v
	}
	return false
}

// emitJSON writes payload to stdout as indented JSON followed by a newline.
// Used as the canonical structured-output writer for success cases.
func emitJSON(cmd *cobra.Command, payload any) error {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return shumerr.Wrap(shumerr.CodeInternal, err, "failed to marshal response")
	}
	if _, err := cmd.OutOrStdout().Write(append(raw, '\n')); err != nil {
		return shumerr.Wrap(shumerr.CodeInternal, err, "failed to write response")
	}
	return nil
}

// EmitError writes a JSON error envelope to stderr.
// Shape: {"error": {"code": "...", "message": "...", "hint": "...", "details": {...}}}
func EmitError(cmd *cobra.Command, e *shumerr.Error) error {
	envelope := struct {
		Error *shumerr.Error `json:"error"`
	}{Error: e}
	raw, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	if _, err := cmd.ErrOrStderr().Write(append(raw, '\n')); err != nil {
		return err
	}
	return nil
}

// encodeJSON is retained as a thin alias so existing call sites keep working
// during the migration. New code should call emitJSON directly.
func encodeJSON(cmd *cobra.Command, payload any) error {
	return emitJSON(cmd, payload)
}
