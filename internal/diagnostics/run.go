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
	if net.ParseIP(t) != nil {
		return true
	}
	return !strings.ContainsAny(t, " /\\;&|`$()<>\"'")
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
