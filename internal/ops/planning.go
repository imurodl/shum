package ops

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type remoteRunner interface {
	Command(alias string, cmd string) (string, error)
}

type PlanningService struct {
	runner remoteRunner
}

func NewPlanningService(runner remoteRunner) *PlanningService {
	return &PlanningService{runner: runner}
}

func (s *PlanningService) Preflight(ctx context.Context, hostAlias string) (PreflightResult, error) {
	select {
	case <-ctx.Done():
		return PreflightResult{}, ctx.Err()
	default:
	}

	checks := map[string]string{
		"docker":          "pending",
		"compose":         "pending",
		"compose_project": "pending",
		"disk":            "unknown",
	}
	out := PreflightResult{
		HostAlias: hostAlias,
		Checks:    checks,
		DiskPath:  "/",
	}

	dockerRaw, err := s.runner.Command(hostAlias, "docker --version")
	if err != nil {
		out.Checks["docker"] = fmt.Sprintf("command failed: %v", err)
		out.Passed = false
		return out, nil
	}
	out.DockerAvailable = true
	out.DockerVersion = normalizeCommandVersion(dockerRaw)
	out.Checks["docker"] = "ok"

	composeRaw, err := s.runner.Command(hostAlias, "docker compose version")
	if err != nil {
		out.Checks["compose"] = fmt.Sprintf("command failed: %v", err)
		out.Passed = false
		return out, nil
	}
	out.ComposeAvailable = true
	out.ComposeVersion = normalizeComposeVersion(composeRaw)
	out.Checks["compose"] = "ok"

	if _, err := s.runner.Command(hostAlias, "docker ps >/dev/null 2>&1"); err != nil {
		out.Checks["compose_project"] = fmt.Sprintf("compose project command failed: %v", err)
		out.PermissionsOK = false
		out.Passed = false
	} else {
		out.PermissionsOK = true
		out.Checks["compose_project"] = "ok"
	}

	diskRaw, err := s.runner.Command(hostAlias, "df -Pk / | awk 'NR==2 {print $4,$6}'")
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(diskRaw))
		if len(parts) >= 2 {
			out.DiskPath = parts[1]
			if available, parseErr := strconv.ParseInt(parts[0], 10, 64); parseErr == nil {
				out.DiskBytesAvail = available * 1024
				out.Checks["disk"] = "ok"
			}
		}
	}
	if out.DiskBytesAvail <= 0 {
		out.Checks["disk"] = "disk check did not return valid values"
	}

	out.Passed = out.DockerAvailable && out.ComposeAvailable && out.PermissionsOK && out.DiskBytesAvail > 0
	return out, nil
}

func normalizeComposeVersion(raw string) string {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return ""
	}
	lines := strings.Split(clean, "\n")
	return strings.TrimSpace(lines[0])
}

func normalizeCommandVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	lines := strings.Split(raw, "\n")
	return strings.TrimSpace(lines[0])
}

func (s *PlanningService) BuildPlan(
	ctx context.Context,
	hostAlias, projectRef string,
	policy ProjectPolicy,
) (Plan, error) {
	if ctx.Err() != nil {
		return Plan{}, ctx.Err()
	}

	preflight, err := s.Preflight(ctx, hostAlias)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{
		HostAlias:  hostAlias,
		ProjectRef: projectRef,
		Preflight:  preflight,
		Policy:     policy,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Actions:    defaultPlanActions(),
		Services:   []ServiceChange{},
	}

	if !preflight.Passed {
		plan.Blocks = append(plan.Blocks, "preflight checks failed")
		return plan, nil
	}

	raw, err := s.runner.Command(hostAlias, "docker compose ps --format json")
	if err != nil {
		plan.Blocks = append(plan.Blocks, fmt.Sprintf("failed to inspect compose services: %v", err))
	} else {
		plan.Services = extractServiceChanges(raw)
		plan.Services = s.fillServiceTargets(ctx, hostAlias, plan.Services)
	}

	if len(plan.Services) == 0 {
		plan.Blocks = append(plan.Blocks, "no discoverable services found in compose status output")
	}

	sort.SliceStable(plan.Services, func(i, j int) bool {
		return plan.Services[i].ServiceName < plan.Services[j].ServiceName
	})

	plan.Warnings = append(plan.Warnings, buildPlanWarnings(preflight, policy, plan.Services)...)
	plan.Blocks = append(plan.Blocks, buildPlanBlocks(preflight, policy, plan.Services)...)
	return plan, nil
}

func buildPlanWarnings(preflight PreflightResult, policy ProjectPolicy, services []ServiceChange) []string {
	warnings := []string{}
	if policy.RequireBackup && policy.BackupCommand == "" {
		warnings = append(warnings, "no backup command configured")
	}
	for _, item := range services {
		if strings.TrimSpace(item.Image) == "" {
			warnings = append(warnings, "some services have unknown image references")
			break
		}
	}
	return uniqueStrings(warnings)
}

func buildPlanBlocks(preflight PreflightResult, policy ProjectPolicy, services []ServiceChange) []string {
	blocks := []string{}
	if !preflight.Passed {
		blocks = append(blocks, "preflight did not pass")
	}
	if len(services) == 0 {
		blocks = append(blocks, "no discoverable services found in compose status output")
	}
	if policy.MigrationWarning {
		blocks = append(blocks, "migration-bearing services detected; explicit confirmation required")
	}
	return blocks
}

func defaultPlanActions() []PlanAction {
	return []PlanAction{
		{
			Name:  "prepare",
			Cmd:   "docker compose pull",
			Notes: "Download newer images before upgrade for deterministic deployment",
		},
		{
			Name:  "apply",
			Cmd:   "docker compose up -d",
			Notes: "Apply new containers and restart services with new images",
		},
	}
}

func (s *PlanningService) fillServiceTargets(ctx context.Context, hostAlias string, services []ServiceChange) []ServiceChange {
	if ctx.Err() != nil {
		return services
	}
	for i := range services {
		img := strings.TrimSpace(services[i].Image)
		if img == "" {
			continue
		}
		currDigest, canRollback, _ := parseDigestFromImage(img)
		services[i].CurrentDigest = currDigest
		services[i].CanRollback = canRollback

		targetDigest := inferTargetDigest(currDigest)
		if targetDigest != "" {
			if remoteDigest, err := s.inspectImageDigest(ctx, hostAlias, img); err == nil {
				targetDigest = remoteDigest
			}
			services[i].TargetDigest = targetDigest
		}
	}
	return services
}

func (s *PlanningService) inspectImageDigest(ctx context.Context, hostAlias, image string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	cmd := fmt.Sprintf("docker image inspect --format '{{index .RepoDigests 0}}' %s", shellEscape(image))
	out, err := s.runner.Command(hostAlias, cmd)
	if err != nil {
		return "", err
	}
	digest := strings.TrimSpace(out)
	if strings.Contains(digest, "@") {
		parts := strings.SplitN(digest, "@", 2)
		return strings.TrimSpace(parts[1]), nil
	}
	return "", fmt.Errorf("image inspect digest unavailable")
}

func parseDigestFromImage(image string) (string, bool, string) {
	if strings.Contains(image, "@") {
		parts := strings.SplitN(image, "@", 2)
		return parts[1], false, parts[0]
	}
	return "", true, image
}

func inferTargetDigest(currentDigest string) string {
	if strings.HasPrefix(currentDigest, "sha256:") {
		return currentDigest
	}
	return currentDigest
}

func shellEscape(raw string) string {
	raw = strings.ReplaceAll(raw, `'`, `'"'"'`)
	return fmt.Sprintf("'%s'", raw)
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
