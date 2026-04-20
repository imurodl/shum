package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/imurodl/shum/internal/shumerr"
)

func TestAgentHelpEmitsValidJSON(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"agent-help"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	if err := root.Execute(); err != nil {
		t.Fatalf("agent-help failed: %v", err)
	}
	var doc agentHelpDoc
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput=%s", err, buf.String())
	}
	if doc.Tool.Name != "shum" {
		t.Errorf("tool.name = %s, want shum", doc.Tool.Name)
	}
	if doc.Contract.JSONFlag != "--json" {
		t.Errorf("contract.json_flag = %s, want --json", doc.Contract.JSONFlag)
	}
	if doc.Contract.ErrorChannel != "stderr" {
		t.Errorf("contract.error_channel = %s, want stderr", doc.Contract.ErrorChannel)
	}
}

func TestAgentHelpIncludesAllErrorCodes(t *testing.T) {
	root := NewRootCommand()
	doc := buildAgentHelp(root)
	for _, code := range shumerr.AllCodes() {
		entry, ok := doc.Errors[code]
		if !ok {
			t.Errorf("error code %q missing from agent-help output", code)
			continue
		}
		if entry.Description == "" {
			t.Errorf("error code %q has empty description", code)
		}
	}
}

func TestAgentHelpListsCoreCommands(t *testing.T) {
	root := NewRootCommand()
	doc := buildAgentHelp(root)
	wantPaths := []string{
		"host register",
		"host list",
		"host inspect",
		"project discover",
		"project inspect",
		"project preflight",
		"project plan",
		"project policy show",
		"project policy set",
		"project backup take",
		"project backup list",
		"project backup restore",
		"project upgrade",
		"project run list",
		"project run show",
		"agent-help",
	}
	got := make(map[string]bool, len(doc.Commands))
	for _, c := range doc.Commands {
		got[c.Path] = true
	}
	for _, path := range wantPaths {
		if !got[path] {
			t.Errorf("command %q missing from agent-help", path)
		}
	}
}

func TestAgentHelpEveryCommandHasShortOrLong(t *testing.T) {
	root := NewRootCommand()
	doc := buildAgentHelp(root)
	for _, c := range doc.Commands {
		if c.Short == "" && c.Long == "" {
			t.Errorf("command %q has neither Short nor Long", c.Path)
		}
	}
}

func TestAgentHelpReturnsHaveShape(t *testing.T) {
	// Every JSON-emitting command should advertise an output shape so agents
	// know what to parse. Commands that don't emit JSON may have empty Returns.
	root := NewRootCommand()
	doc := buildAgentHelp(root)
	jsonCommands := map[string]bool{
		"host register":          true,
		"host list":              true,
		"host inspect":           true,
		"project discover":       true,
		"project inspect":        true,
		"project preflight":      true,
		"project plan":           true,
		"project policy show":    true,
		"project backup take":    true,
		"project backup list":    true,
		"project upgrade":        true,
		"project run list":       true,
		"project run show":       true,
		"agent-help":             true,
	}
	for _, c := range doc.Commands {
		if jsonCommands[c.Path] && strings.TrimSpace(c.Returns) == "" {
			t.Errorf("json-emitting command %q has empty Returns shape", c.Path)
		}
	}
}
