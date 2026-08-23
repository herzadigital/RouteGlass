package juniper

import (
	"context"
	"errors"
	"fmt"
	"github.com/herzadigital/routeglass/internal/config"
	"github.com/herzadigital/routeglass/internal/mikrotik"
	"net/netip"
	"os/exec"
	"regexp"
	"strings"
)

type Client struct{ cfg config.JuniperConfig }

var ErrDisabled = errors.New("BGP lookup is not enabled")

func New(c config.JuniperConfig) *Client { return &Client{c} }

var prefixRE = regexp.MustCompile(`^([0-9A-Fa-f:.]+/[0-9]+)\s+\*?\[([^/\]]+)/([0-9]+)\]`)
var asRE = regexp.MustCompile(`AS path:\s*(.*?)\s*(?:,|$)`)
var nhRE = regexp.MustCompile(`^\s*>\s+to\s+([^\s]+)`)

func (c *Client) Lookup(ctx context.Context, target string) (mikrotik.Result, error) {
	ip, err := netip.ParseAddr(target)
	if err != nil {
		return mikrotik.Result{}, errors.New("target must be a valid IPv4 or IPv6 address")
	}
	if c.cfg.Address == "" || c.cfg.Username == "" || c.cfg.PrivateKeyFile == "" {
		return mikrotik.Result{}, errors.New("invalid Juniper configuration")
	}
	cmd := exec.CommandContext(ctx, "ssh", "-i", c.cfg.PrivateKeyFile, "-p", fmt.Sprint(c.cfg.Port), "-o", "BatchMode=yes", "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=yes", "-o", "UserKnownHostsFile="+c.cfg.KnownHostsFile, "-o", "ConnectTimeout=5", c.cfg.Username+"@"+c.cfg.Address, "show route "+ip.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return mikrotik.Result{}, fmt.Errorf("Juniper lookup failed: %s", strings.TrimSpace(string(out)))
	}
	lines := strings.Split(string(out), "\n")
	res := mikrotik.Result{Target: ip.String()}
	active := false
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if m := prefixRE.FindStringSubmatch(t); m != nil {
			if strings.Contains(m[0], "*") {
				active = true
				res.Prefix = m[1]
				res.Protocol = m[2]
				res.Distance = m[3]
			} else if res.Prefix == "" {
				res.Prefix = m[1]
				res.Protocol = m[2]
				res.Distance = m[3]
			}
		}
		if m := asRE.FindStringSubmatch(t); m != nil && (active || res.ASPath == "") {
			res.ASPath = strings.TrimSpace(m[1])
		}
		if m := nhRE.FindStringSubmatch(ln); m != nil && (active || res.Gateway == "") {
			res.Gateway = m[1]
		}
	}
	return res, nil
}
