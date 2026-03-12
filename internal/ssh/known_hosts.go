package ssh

import (
	"crypto/sha256"
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
)

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
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		keyData := fields[2]
		decoded, err := base64.StdEncoding.DecodeString(keyData)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(decoded)
		return "SHA256:" + base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(sum[:])
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
	}
	return ""
}
