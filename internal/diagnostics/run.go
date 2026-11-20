package diagnostics

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

func validTarget(t string) bool {
	if t == "" || len(t) > 253 {
		return false
	}
	if ip := net.ParseIP(t); ip != nil {
		return true
	}
	// Targets are passed to external diagnostic programs. Restrict hostnames
	// to DNS labels so user input cannot be interpreted as command options.
	// In particular, a leading dash must never reach ping/mtr/traceroute.
	if strings.HasSuffix(t, ".") {
		t = strings.TrimSuffix(t, ".")
	}
	if t == "" {
		return false
	}
	for _, label := range strings.Split(t, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '-') {
				return false
			}
		}
	}
	return true
}
func Run(kind, target string) (string, error) {
	if !validTarget(target) {
		return "", fmt.Errorf("invalid target")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	switch kind {
	case "ping":
		cmd = exec.CommandContext(ctx, "/usr/bin/ping", "-c", "5", "-W", "3", target)
	case "mtr":
		cmd = exec.CommandContext(ctx, "/usr/bin/mtr", "-r", "-c", "5", "-w", target)
	case "traceroute":
		cmd = exec.CommandContext(ctx, "/usr/bin/traceroute", "-n", "-w", "3", "-q", "2", target)
	default:
		return "", fmt.Errorf("unsupported diagnostic")
	}
	out, e := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("command timeout")
	}
	return string(out), e
}
