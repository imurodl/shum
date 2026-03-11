package ssh

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type ResolvedConfig struct {
	Alias       string
	User        string
	Hostname    string
	Port        int
	IdentityFiles []string
	KnownHostFiles []string
}

func ParseResolvedAlias(alias string) (*ResolvedConfig, error) {
	out, err := exec.Command("ssh", "-G", alias).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ssh alias: %w", err)
	}

	parsed := map[string][]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		parsed[key] = append(parsed[key], val)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	cfg := &ResolvedConfig{
		Alias: alias,
		Port:  22,
	}

	if host := strings.TrimSpace(firstOrEmpty(parsed["hostname"])); host != "" {
		cfg.Hostname = host
	} else {
		cfg.Hostname = alias
	}
	if usr := strings.TrimSpace(firstOrEmpty(parsed["user"])); usr != "" {
		cfg.User = usr
	}
	cfg.IdentityFiles = uniqueStrings(parsed["identityfile"])
	cfg.KnownHostFiles = uniqueStrings(append(parsed["userknownhostsfile"], parsed["globalknownhostsfile"]...))

	if raw := firstOrEmpty(parsed["port"]); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid ssh port: %w", err)
		}
		cfg.Port = port
	}
	if cfg.User == "" {
		cfg.User = ""
	}
	return cfg, nil
}

func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
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
