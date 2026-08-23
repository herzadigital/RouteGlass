package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type MikroTikConfig struct {
	Enabled        bool   `json:"enabled"`
	Address        string `json:"address"`
	Username       string `json:"username"`
	PasswordFile   string `json:"password_file"`
	CAFile         string `json:"ca_file"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}
type JuniperConfig struct {
	Enabled        bool   `json:"enabled"`
	Address        string `json:"address"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	PrivateKeyFile string `json:"private_key_file"`
	KnownHostsFile string `json:"known_hosts_file"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}
type Config struct {
	CompanyName string         `json:"company_name"`
	ASN         string         `json:"asn"`
	Location    string         `json:"location"`
	TestIPv4    string         `json:"test_ipv4"`
	TestIPv6    string         `json:"test_ipv6"`
	TestFiles   []int          `json:"test_files"`
	RateLimit   bool           `json:"rate_limit"`
	MikroTik    MikroTikConfig `json:"mikrotik"`
	Juniper     JuniperConfig  `json:"juniper"`
}

func Load(p string) (Config, error) {
	var c Config
	b, e := os.ReadFile(p)
	if e != nil {
		return c, fmt.Errorf("read config: %w", e)
	}
	if e = json.Unmarshal(b, &c); e != nil {
		return c, fmt.Errorf("parse config: %w", e)
	}
	c.MikroTik.Address = strings.TrimSpace(c.MikroTik.Address)
	c.Juniper.Address = strings.TrimSpace(c.Juniper.Address)
	if c.MikroTik.TimeoutSeconds <= 0 {
		c.MikroTik.TimeoutSeconds = 8
	}
	if c.Juniper.TimeoutSeconds <= 0 {
		c.Juniper.TimeoutSeconds = 8
	}
	if c.Juniper.Port <= 0 {
		c.Juniper.Port = 22
	}
	return c, nil
}
