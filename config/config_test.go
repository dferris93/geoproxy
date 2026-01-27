package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfigSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`apiKey: "abc"
servers:
  - listenIP: "127.0.0.1"
    listenPort: "8080"
    backendIP: "10.0.0.1"
    backendPort: "9090"
    allowedCountries: ["US", "CA"]
    recvProxyProtocol: true
    proxyProtocolVersion: 2
    daysOfWeek: ["Monday", "Tuesday"]
    startDate: "2024-01-01"
    endDate: "2024-12-31"
    startTime: "09:00"
    endTime: "17:00"
`)

	err := os.WriteFile(path, content, 0o600)
	assert.NoError(t, err)

	cfg, err := ReadConfig(path)
	assert.NoError(t, err)
	if assert.NotNil(t, cfg) {
		assert.Equal(t, "abc", cfg.APIKey)
		assert.Len(t, cfg.Servers, 1)
		assert.Equal(t, "127.0.0.1", cfg.Servers[0].ListenIP)
		assert.Equal(t, "8080", cfg.Servers[0].ListenPort)
		assert.Equal(t, "10.0.0.1", cfg.Servers[0].BackendIP)
		assert.Equal(t, "9090", cfg.Servers[0].BackendPort)
		assert.Equal(t, []string{"US", "CA"}, cfg.Servers[0].AllowedCountries)
		assert.True(t, cfg.Servers[0].RecvProxyProtocol)
		assert.Equal(t, 2, cfg.Servers[0].ProxyProtocolVersion)
	}
}

func TestReadConfigMissingFile(t *testing.T) {
	_, err := ReadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
}

func TestReadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte("servers: [:"), 0o600)
	assert.NoError(t, err)

	_, err = ReadConfig(path)
	assert.Error(t, err)
}

func TestValidateTrustedProxies(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		wantErr bool
	}{
		{name: "empty", entries: nil, wantErr: false},
		{name: "valid ip", entries: []string{"192.0.2.1"}, wantErr: false},
		{name: "valid cidr", entries: []string{"192.0.2.0/24"}, wantErr: false},
		{name: "invalid ip", entries: []string{"999.0.0.1"}, wantErr: true},
		{name: "invalid cidr", entries: []string{"192.0.2.0/33"}, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateTrustedProxies(tc.entries)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
