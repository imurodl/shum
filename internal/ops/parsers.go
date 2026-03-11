package ops

import (
	"encoding/json"
	"fmt"
	"strings"
)

type composePSLine struct {
	Service string `json:"Service"`
	Name    string `json:"Name"`
	Image   string `json:"Image"`
	State   string `json:"State"`
	Health  string `json:"Health"`
}

// extractServiceChanges parses `docker compose ps --format json` output into
// ServiceChange entries. Name (container name, e.g. "web-1") takes priority
// over Service (compose service name, e.g. "web") because Name is always
// unique; older Compose versions may only populate one of the two fields.
func extractServiceChanges(raw string) []ServiceChange {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []ServiceChange{}
	}

	var rows []composePSLine
	if err := parseComposePS(raw, &rows); err != nil {
		return []ServiceChange{}
	}

	out := make([]ServiceChange, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" {
			name = strings.TrimSpace(row.Service)
		}
		change := ServiceChange{
			ServiceName:    name,
			Image:          strings.TrimSpace(row.Image),
			CurrentHealthy: chooseHealth(row.State, row.Health),
		}
		if change.ServiceName == "" {
			continue
		}
		out = append(out, change)
	}
	return out
}

// parseComposePS handles three Docker Compose output formats for
// `docker compose ps --format json`:
//
// Format 1 — JSON array (Docker Compose v2.21+):
//
//	[{"Service":"web","Image":"nginx:1.27","State":"running","Health":"healthy"}, ...]
//
// Format 2 — Newline-delimited JSON objects (Docker Compose v2.0–v2.20):
//
//	{"Name":"web-1","Image":"nginx:1.27","State":"running"}
//	{"Name":"api-1","Image":"app:latest","State":"running"}
//
// Format 3 — JSON object keyed by service name (some Compose v1 plugins):
//
//	{"web":{"Service":"web","Image":"nginx","State":"running","Health":"healthy"}}
//
// The parser tries Format 1 first (input starts with '['), then falls back to
// line-by-line parsing which handles both Format 2 and Format 3. This ordering
// ensures compatibility across Docker Compose versions without requiring the
// user to specify which version they run.
func parseComposePS(raw string, out *[]composePSLine) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	if strings.HasPrefix(raw, "[") {
		var rows []composePSLine
		if err := json.Unmarshal([]byte(raw), &rows); err != nil {
			// fallback: object map keyed by service name.
			var obj map[string]any
			if err := json.Unmarshal([]byte(raw), &obj); err != nil {
				return err
			}
			rows = append(rows, entriesFromMap(obj)...)
		}
		if len(rows) > 0 {
			*out = rows
			return nil
		}
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row composePSLine
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			if strings.TrimSpace(row.Name) != "" || strings.TrimSpace(row.Service) != "" ||
				strings.TrimSpace(row.Image) != "" || strings.TrimSpace(row.State) != "" || strings.TrimSpace(row.Health) != "" {
				*out = append(*out, row)
				continue
			}
		}
		var generic map[string]any
		if err := json.Unmarshal([]byte(line), &generic); err == nil {
			*out = append(*out, entriesFromMap(generic)...)
			continue
		}
		return fmt.Errorf("unable to parse compose status line: %s", line)
	}

	return nil
}

func entriesFromMap(raw map[string]any) []composePSLine {
	out := []composePSLine{}
	for _, value := range raw {
		obj, ok := value.(map[string]any)
		if !ok {
			continue
		}
		line := composePSLine{
			Service: castToString(obj["Service"]),
			Name:    castToString(obj["Name"]),
			Image:   castToString(obj["Image"]),
			State:   castToString(obj["State"]),
			Health:  castToString(obj["Health"]),
		}
		if strings.TrimSpace(line.Name) == "" {
			line.Name = castToString(obj["name"])
		}
		if strings.TrimSpace(line.Service) == "" {
			line.Service = castToString(obj["service"])
		}
		if strings.TrimSpace(line.State) == "" {
			line.State = castToString(obj["state"])
		}
		if strings.TrimSpace(line.Health) == "" {
			line.Health = castToString(obj["health"])
		}
		if strings.TrimSpace(line.Image) == "" {
			line.Image = castToString(obj["image"])
		}
		if strings.TrimSpace(line.Name) == "" && strings.TrimSpace(line.Service) == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func castToString(raw any) string {
	if raw == nil {
		return ""
	}
	switch val := raw.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		return ""
	}
}

func chooseHealth(values ...string) string {
	for _, value := range values {
		clean := strings.TrimSpace(strings.ToLower(value))
		if clean != "" {
			return clean
		}
	}
	return ""
}
