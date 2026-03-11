package ops

import (
	"fmt"
	"strings"
)

type mockResponse struct {
	output string
	err    error
}

type mockRunner struct {
	responses map[string]mockResponse
	calls     []string
}

func newMockRunner() *mockRunner {
	return &mockRunner{responses: make(map[string]mockResponse)}
}

func (m *mockRunner) on(cmdSubstring, output string, err error) {
	m.responses[cmdSubstring] = mockResponse{output: output, err: err}
}

func (m *mockRunner) Command(alias string, cmd string) (string, error) {
	m.calls = append(m.calls, cmd)
	for pattern, resp := range m.responses {
		if strings.Contains(cmd, pattern) {
			return resp.output, resp.err
		}
	}
	return "", nil
}

// preflightOK sets up mock responses for a passing preflight check.
func (m *mockRunner) preflightOK() {
	m.on("docker --version", "Docker version 27.0.0", nil)
	m.on("docker compose version", "Docker Compose version v2.28.0", nil)
	m.on("docker ps", "", nil)
	m.on("df -Pk", "50000000 /", nil)
}

// composePSOK sets up mock response for docker compose ps returning one healthy service.
func (m *mockRunner) composePSOK() {
	m.on("docker compose ps --format json", `[{"Service":"web","Name":"web-1","Image":"nginx:1.27","State":"running","Health":"healthy"}]`, nil)
}

// imageInspectOK sets up mock response for image digest resolution.
func (m *mockRunner) imageInspectOK() {
	m.on("docker image inspect", "sha256:abc123", nil)
}

// upgradeOK sets up mock responses for a successful upgrade (pull + up).
func (m *mockRunner) upgradeOK() {
	m.on("docker compose pull", "", nil)
	m.on("docker compose up -d", "", nil)
}

// upgradeOKWithBackup sets up mock responses for a successful upgrade with backup.
func (m *mockRunner) upgradeOKWithBackup() {
	m.upgradeOK()
	m.on("SHUM_BACKUP_ARTIFACT=", "backup data", nil)
}

// pullFails sets up mock response for a failed compose pull.
func (m *mockRunner) pullFails() {
	m.on("docker compose pull", "", fmt.Errorf("network timeout"))
}
