package config

import (
	"os"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Servers []ServerConfig `yaml:"servers"`
}

type ServerConfig struct {
	ListenIP         string   `yaml:"listenIP"`
	ListenPort       string   `yaml:"listenPort"`
	BackendIP        string   `yaml:"backendIP"`
	BackendPort      string   `yaml:"backendPort"`
	AllowedCountries []string `yaml:"allowedCountries"`
	AllowedRegions   []string `yaml:"allowedRegions"`
	AlwaysAllowed 	 []string `yaml:"alwaysAllowed"`
	AlwaysDenied     []string `yaml:"alwaysDenied"`
	DeniedCountries  []string `yaml:"deniedCountries"`
	DeniedRegions    []string `yaml:"deniedRegions"`
	UseProxyProtocol bool     `yaml:"useProxyProtocol"`
	ProxyProtocolVersion int  `yaml:"proxyProtocolVersion"`
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