package mikrotik

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/herzadigital/routeglass/internal/config"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"
)

type Result struct {
	Target   string            `json:"target"`
	Prefix   string            `json:"prefix,omitempty"`
	Gateway  string            `json:"gateway,omitempty"`
	Protocol string            `json:"protocol,omitempty"`
	ASPath   string            `json:"as_path,omitempty"`
	Comment  string            `json:"comment,omitempty"`
	Distance string            `json:"distance,omitempty"`
	Raw      map[string]string `json:"raw,omitempty"`
}
type Client struct{ cfg config.MikroTikConfig }

var ErrDisabled = errors.New("BGP lookup is not enabled")

func New(c config.MikroTikConfig) *Client { return &Client{c} }
func encLen(n int) []byte {
	switch {
	case n < 0x80:
		return []byte{byte(n)}
	case n < 0x4000:
		return []byte{byte((n >> 8) | 0x80), byte(n)}
	case n < 0x200000:
		return []byte{byte((n >> 16) | 0xC0), byte(n >> 8), byte(n)}
	case n < 0x10000000:
		return []byte{byte((n >> 24) | 0xE0), byte(n >> 16), byte(n >> 8), byte(n)}
	default:
		return []byte{0xF0, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
	}
}
func decLen(r io.ByteReader) (int, error) {
	b, e := r.ReadByte()
	if e != nil {
		return 0, e
	}
	switch {
	case b < 0x80:
		return int(b), nil
	case b < 0xC0:
		c, e := r.ReadByte()
		return (int(b&0x3F) << 8) | int(c), e
	case b < 0xE0:
		c, e := r.ReadByte()
		if e != nil {
			return 0, e
		}
		d, e := r.ReadByte()
		return (int(b&0x1F) << 16) | (int(c) << 8) | int(d), e
	case b < 0xF0:
		c, e := r.ReadByte()
		if e != nil {
			return 0, e
		}
		d, e := r.ReadByte()
		if e != nil {
			return 0, e
		}
		f, e := r.ReadByte()
		return (int(b&0x0F) << 24) | (int(c) << 16) | (int(d) << 8) | int(f), e
	default:
		x := 0
		for i := 0; i < 4; i++ {
			c, e := r.ReadByte()
			if e != nil {
				return 0, e
			}
			x = x<<8 | int(c)
		}
		return x, nil
	}
}
func writeSentence(w io.Writer, words ...string) error {
	for _, s := range words {
		if _, e := w.Write(encLen(len(s))); e != nil {
			return e
		}
		if _, e := io.WriteString(w, s); e != nil {
			return e
		}
	}
	_, e := w.Write([]byte{0})
	return e
}
func readSentence(r *bufio.Reader) ([]string, error) {
	var out []string
	for {
		n, e := decLen(r)
		if e != nil {
			return nil, e
		}
		if n == 0 {
			return out, nil
		}
		b := make([]byte, n)
		if _, e = io.ReadFull(r, b); e != nil {
			return nil, e
		}
		out = append(out, string(b))
	}
}

func parse(words []string) map[string]string {
	m := map[string]string{}
	for _, w := range words {
		if strings.HasPrefix(w, "=") {
			x := strings.SplitN(w[1:], "=", 2)
			if len(x) == 2 {
				m[x[0]] = x[1]
			}
		}
	}
	return m
}
func (c *Client) dial(ctx context.Context) (*tls.Conn, error) {
	pem, e := os.ReadFile(c.cfg.CAFile)
	if e != nil {
		return nil, e
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("invalid MikroTik CA certificate")
	}
	d := net.Dialer{}
	raw, e := d.DialContext(ctx, "tcp", c.cfg.Address)
	if e != nil {
		return nil, e
	}
	host, _, e := net.SplitHostPort(c.cfg.Address)
	if e != nil {
		host = c.cfg.Address
	}
	tc := tls.Client(raw, &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12, ServerName: host})
	if e = tc.HandshakeContext(ctx); e != nil {
		raw.Close()
		return nil, e
	}
	return tc, nil
}
func trapError(words []string) error {
	m := parse(words[1:])
	if msg := m["message"]; msg != "" {
		return fmt.Errorf("MikroTik API error: %s", msg)
	}
	return errors.New("MikroTik API request rejected")
}

func prefixFor(row map[string]string) (netip.Prefix, bool) {
	p := row["dst-address"]
	if p == "" {
		p = row["dst"]
	}
	if p == "" {
		return netip.Prefix{}, false
	}
	v, err := netip.ParsePrefix(p)
	if err != nil {
		return netip.Prefix{}, false
	}
	return v.Masked(), true
}

func (c *Client) Lookup(ctx context.Context, target string) (Result, error) {
	ip, err := netip.ParseAddr(target)
	if err != nil {
		return Result{}, errors.New("target must be a valid IPv4 or IPv6 address")
	}
	pw, err := os.ReadFile(c.cfg.PasswordFile)
	if err != nil {
		return Result{}, err
	}
	tc, err := c.dial(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tc.Close()
	br := bufio.NewReader(tc)

	if err = writeSentence(tc, "/login", "=name="+c.cfg.Username, "=password="+strings.TrimSpace(string(pw))); err != nil {
		return Result{}, err
	}
	for {
		words, err := readSentence(br)
		if err != nil {
			return Result{}, err
		}
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "!trap":
			return Result{}, trapError(words)
		case "!done":
			goto loggedIn
		}
	}
loggedIn:

	if err = writeSentence(tc, "/routing/route/print"); err != nil {
		return Result{}, err
	}

	var best map[string]string
	bestBits := -1
	for {
		words, err := readSentence(br)
		if err != nil {
			return Result{}, err
		}
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "!re":
			row := parse(words[1:])
			pfx, ok := prefixFor(row)
			if !ok || !pfx.Contains(ip) {
				continue
			}
			if pfx.Bits() > bestBits {
				best = row
				bestBits = pfx.Bits()
			}
		case "!trap":
			return Result{}, trapError(words)
		case "!done":
			goto finished
		}
	}
finished:

	if best == nil {
		return Result{Target: ip.String()}, nil
	}
	prefix := best["dst-address"]
	if prefix == "" {
		prefix = best["dst"]
	}
	asPath := best["bgp.as-path"]
	if asPath == "" {
		asPath = best["as-path"]
	}
	protocol := best["protocol"]
	if protocol == "" && best["belongs-to"] != "" {
		protocol = best["belongs-to"]
	}
	return Result{
		Target: ip.String(),
		Prefix: prefix,
		Gateway: best["gateway"],
		Protocol: protocol,
		ASPath: asPath,
		Comment: best["comment"],
		Distance: best["distance"],
		Raw: best,
	}, nil
}

var _ = time.Second
