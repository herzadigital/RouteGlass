package vyos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/herzadigital/routeglass/internal/config"
	"github.com/herzadigital/routeglass/internal/mikrotik"
)

type Client struct{ cfg config.VyOSConfig }

func New(c config.VyOSConfig) *Client { return &Client{c} }

func (c *Client) run(ctx context.Context, command string) ([]byte, error) {
	if c.cfg.Address == "" || c.cfg.Username == "" || c.cfg.PrivateKeyFile == "" || c.cfg.KnownHostsFile == "" {
		return nil, errors.New("invalid VyOS configuration")
	}

	args := []string{
		"-6",
		"-i", c.cfg.PrivateKeyFile,
		"-p", fmt.Sprint(c.cfg.Port),
		"-o", "BatchMode=yes",
		"-o", "IdentitiesOnly=yes",
		"-o", "StrictHostKeyChecking=yes",
		"-o", "UserKnownHostsFile=" + c.cfg.KnownHostsFile,
		"-o", "ConnectTimeout=5",
		c.cfg.Username + "@" + c.cfg.Address,
		command,
	}

	cmd := exec.CommandContext(ctx, "ssh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"VyOS lookup failed: %s",
			strings.TrimSpace(string(out)),
		)
	}

	return out, nil
}

func routeCommand(ip netip.Addr) string {
	if ip.Is6() {
		return `vtysh -c "show ipv6 route ` + ip.String() + ` json"`
	}
	return `vtysh -c "show ip route ` + ip.String() + ` json"`
}

func bgpCommand(pfx string, ipv6 bool) string {
	if ipv6 {
		return `vtysh -c "show bgp ipv6 unicast ` + pfx + ` json"`
	}
	return `vtysh -c "show bgp ipv4 unicast ` + pfx + ` json"`
}

func firstString(v any, keys ...string) string {
	if m, ok := v.(map[string]any); ok {
		for _, k := range keys {
			if x, ok := m[k]; ok {
				if s, ok := x.(string); ok && s != "" {
					return s
				}
				if f, ok := x.(float64); ok {
					return fmt.Sprintf("%g", f)
				}
			}
		}
	}
	return ""
}
func findPrefix(v any, ip netip.Addr) (string, map[string]any) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if p, err := netip.ParsePrefix(k); err == nil && p.Contains(ip) {
				if m, ok := child.(map[string]any); ok {
					return p.String(), m
				}
				if a, ok := child.([]any); ok && len(a) > 0 {
					if m, ok := a[0].(map[string]any); ok {
						return p.String(), m
					}
				}
			}
		}
		for _, child := range x {
			if p, m := findPrefix(child, ip); p != "" {
				return p, m
			}
		}
	case []any:
		for _, child := range x {
			if p, m := findPrefix(child, ip); p != "" {
				return p, m
			}
		}
	}
	return "", nil
}
func pathInfo(v any) (asPath, nextHop string) {
	root, _ := v.(map[string]any)
	paths, _ := root["paths"].([]any)
	if len(paths) == 0 {
		return "", ""
	}
	chosen, _ := paths[0].(map[string]any)
	for _, p := range paths {
		if m, ok := p.(map[string]any); ok {
			if bp, ok := m["bestpath"].(map[string]any); ok {
				if overall, _ := bp["overall"].(bool); overall {
					chosen = m
					break
				}
			}
		}
	}
	if ap, ok := chosen["aspath"].(map[string]any); ok {
		asPath = firstString(ap, "string")
	}
	if nhs, ok := chosen["nexthops"].([]any); ok && len(nhs) > 0 {
		nextHop = firstString(nhs[0], "ip")
	}
	return
}

func (c *Client) Lookup(ctx context.Context, target string) (mikrotik.Result, error) {
	ip, err := netip.ParseAddr(target)
	if err != nil {
		return mikrotik.Result{}, errors.New("target must be a valid IPv4 or IPv6 address")
	}
	out, err := c.run(ctx, routeCommand(ip))
	if err != nil {
		return mikrotik.Result{}, err
	}
	var route any
	if err = json.Unmarshal(out, &route); err != nil {
		return mikrotik.Result{}, fmt.Errorf("VyOS returned invalid route JSON: %w", err)
	}
	pfx, routeInfo := findPrefix(route, ip)
	res := mikrotik.Result{Target: ip.String()}
	if pfx == "" {
		return res, nil
	}
	res.Prefix = pfx
	res.Protocol = firstString(routeInfo, "protocol", "type")
	res.Distance = firstString(routeInfo, "distance", "adminDistance")
	res.Gateway = firstString(routeInfo, "nexthop", "nextHop", "gateway")
	out, err = c.run(ctx, bgpCommand(pfx, ip.Is6()))
	if err == nil {
		var bgp any
		if json.Unmarshal(out, &bgp) == nil {
			asPath, nh := pathInfo(bgp)
			res.ASPath = asPath
			if nh != "" {
				res.Gateway = nh
			}
			if res.Protocol == "" {
				res.Protocol = "bgp"
			}
		}
	}
	return res, nil
}
