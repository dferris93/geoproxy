package config

import (
	"gopkg.in/yaml.v2"
	"os"
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
	StartTime            string   `yaml:"startTime"`
	EndTime              string   `yaml:"endTime"`
}

func ReadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
