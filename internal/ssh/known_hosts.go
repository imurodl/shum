package ssh

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var fingerprintRegex = regexp.MustCompile(`\s([A-Za-z0-9+/=]+)\s`)

func VerifyHostKey(host string, port int, knownHostFiles []string) (string, error) {
	targets := []string{host}
	if port != 22 {
		targets = append(targets, fmt.Sprintf("[%s]:%d", host, port))
	}

	for _, file := range knownHostFiles {
		for _, target := range targets {
			cmd := exec.Command("ssh-keygen", "-F", target, "-f", file)
			out, err := cmd.Output()
			if err == nil && len(out) > 0 {
				fingerprint := extractFingerprint(out)
				if fingerprint == "" {
					fingerprint = strings.TrimSpace(string(bytes.TrimSpace(out)))
				}
				return fingerprint, nil
			}
		}
	}
	return "", fmt.Errorf("strict host key verification failed for %s", host)
}

func extractFingerprint(output []byte) string {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := fingerprintRegex.FindStringSubmatch(line)
		if len(m) == 2 {
			return m[1]
		}
	}
	return ""
}
