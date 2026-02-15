package config

import (
	"fmt"
	"net"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Servers []ServerConfig `yaml:"servers"`
	APIKey  string         `yaml:"apiKey"`
}

type ServerConfig struct {
	ListenIP             string   `yaml:"listenIP"`
	ListenPort           string   `yaml:"listenPort"`
	BackendIP            string   `yaml:"backendIP"`
	BackendPort          string   `yaml:"backendPort"`
	AllowedCountries     []string `yaml:"allowedCountries"`
	AllowedRegions       []string `yaml:"allowedRegions"`
	AlwaysAllowed        []string `yaml:"alwaysAllowed"`
	AlwaysDenied         []string `yaml:"alwaysDenied"`
	DeniedCountries      []string `yaml:"deniedCountries"`
	DeniedRegions        []string `yaml:"deniedRegions"`
	RecvProxyProtocol    bool     `yaml:"recvProxyProtocol"`
	SendProxyProtocol    bool     `yaml:"sendProxyProtocol"`
	ProxyProtocolVersion int      `yaml:"proxyProtocolVersion"`
	TrustedProxies       []string `yaml:"trustedProxies"`
	DaysOfWeek           []string `yaml:"daysOfWeek"`
	StartDate            string   `yaml:"startDate"`
	EndDate              string   `yaml:"endDate"`
	StartTime            string   `yaml:"startTime"`
	EndTime              string   `yaml:"endTime"`
}

func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		return nil, err
	}

	for i, server := range config.Servers {
		if err := validateTrustedProxies(server.TrustedProxies); err != nil {
			return nil, fmt.Errorf("server %d trustedProxies: %w", i, err)
		}
		if err := validateIPOrCIDREntries(server.AlwaysAllowed); err != nil {
			return nil, fmt.Errorf("server %d alwaysAllowed: %w", i, err)
		}
		if err := validateIPOrCIDREntries(server.AlwaysDenied); err != nil {
			return nil, fmt.Errorf("server %d alwaysDenied: %w", i, err)
		}
	}

	return &config, nil
}

func validateTrustedProxies(entries []string) error {
	for _, entry := range entries {
		if entry == "" {
			return fmt.Errorf("invalid IP %q", entry)
		}
		if _, _, err := net.ParseCIDR(entry); err == nil {
			return fmt.Errorf("CIDRs are not allowed in trustedProxies (got %q); use a plain IPv4/IPv6 address", entry)
		}
		if net.ParseIP(entry) == nil {
			return fmt.Errorf("invalid IP %q", entry)
		}
	}
	return nil
}

func validateIPOrCIDREntries(entries []string) error {
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			return fmt.Errorf("invalid IP/CIDR %q", entry)
		}
		if net.ParseIP(entry) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(entry); err == nil {
			continue
		}
		return fmt.Errorf("invalid IP/CIDR %q", entry)
	}
	return nil
}
