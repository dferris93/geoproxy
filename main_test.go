package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"geoproxy/ipapi"
	"geoproxy/server"
)

type startCapture struct {
	calls   int
	configs []server.ServerConfig
}

func (s *startCapture) start(cfg server.ServerConfig, wg *sync.WaitGroup, _ context.Context) {
	s.calls++
	s.configs = append(s.configs, cfg)
	wg.Done()
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestParseWeekdayVariants(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Weekday
	}{
		{name: "sun", in: "Sun", want: time.Sunday},
		{name: "mon", in: "monday", want: time.Monday},
		{name: "tue", in: "Tues", want: time.Tuesday},
		{name: "wed", in: "WEDNESDAY", want: time.Wednesday},
		{name: "thu", in: "thurs", want: time.Thursday},
		{name: "fri", in: "Friday", want: time.Friday},
		{name: "sat", in: "sat", want: time.Saturday},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseWeekday(tc.in)
			if err != nil {
				t.Fatalf("parseWeekday: %v", err)
			}
			if got != tc.want {
				t.Fatalf("parseWeekday(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}

	if _, err := parseWeekday("nope"); err == nil {
		t.Fatal("expected error for invalid weekday")
	}
}

func TestParseDaysOfWeek(t *testing.T) {
	got, err := parseDaysOfWeek(nil)
	if err != nil {
		t.Fatalf("parseDaysOfWeek: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}

	got, err = parseDaysOfWeek([]string{"Mon", "Wednesday"})
	if err != nil {
		t.Fatalf("parseDaysOfWeek: %v", err)
	}
	if !got[time.Monday] || !got[time.Wednesday] {
		t.Fatalf("expected monday and wednesday set, got %v", got)
	}

	if _, err := parseDaysOfWeek([]string{"nope"}); err == nil {
		t.Fatal("expected error for invalid day")
	}
}

func TestRunHappyPathDateRange(t *testing.T) {
	path := writeConfig(t, `apiKey: "abc"
servers:
  - listenIP: "127.0.0.1"
    listenPort: "8000"
    backendIP: "127.0.0.1"
    backendPort: "9000"
    allowedCountries: ["US"]
    recvProxyProtocol: false
    sendProxyProtocol: true
    proxyProtocolVersion: 2
    trustedProxies: ["10.0.0.1"]
    startDate: "2024-01-01"
    endDate: "2024-01-02"
    startTime: "08:00"
    endTime: "17:00"
`)

	capture := &startCapture{}
	err := run([]string{"-config", path}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: capture.start,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if capture.calls != 1 {
		t.Fatalf("expected 1 startServer call, got %d", capture.calls)
	}

	cfg := capture.configs[0]
	if cfg.TrustedProxies != nil {
		t.Fatalf("expected trusted proxies to be nil when recvProxyProtocol is false")
	}

	factory, ok := cfg.HandlerFactory.(*server.HandlerFactory)
	if !ok {
		t.Fatalf("expected HandlerFactory to be *server.HandlerFactory, got %T", cfg.HandlerFactory)
	}
	if !factory.SendProxyProtocol || factory.ProxyProtocolVersion != 2 {
		t.Fatalf("unexpected proxy protocol settings: send=%v version=%d", factory.SendProxyProtocol, factory.ProxyProtocolVersion)
	}
	if factory.StartDate.IsZero() || factory.EndDate.IsZero() {
		t.Fatal("expected start/end dates to be parsed")
	}
	if factory.StartTime.IsZero() || factory.EndTime.IsZero() {
		t.Fatal("expected start/end times to be parsed")
	}
	if len(factory.DaysOfWeek) != 0 {
		t.Fatalf("expected no days of week, got %v", factory.DaysOfWeek)
	}
}

func TestRunWithDaysOfWeek(t *testing.T) {
	path := writeConfig(t, `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8001"
    backendIP: "127.0.0.1"
    backendPort: "9001"
    allowedCountries: ["US"]
    recvProxyProtocol: true
    trustedProxies: ["10.0.0.1"]
    daysOfWeek: ["Mon", "Wednesday"]
    startTime: "07:00"
    endTime: "19:00"
`)

	capture := &startCapture{}
	err := run([]string{"-config", path}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: capture.start,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if capture.calls != 1 {
		t.Fatalf("expected 1 startServer call, got %d", capture.calls)
	}
	cfg := capture.configs[0]
	if len(cfg.TrustedProxies) != 1 || cfg.TrustedProxies[0] != "10.0.0.1" {
		t.Fatalf("expected trusted proxies to be preserved, got %v", cfg.TrustedProxies)
	}
	factory, ok := cfg.HandlerFactory.(*server.HandlerFactory)
	if !ok {
		t.Fatalf("expected HandlerFactory to be *server.HandlerFactory, got %T", cfg.HandlerFactory)
	}
	if !factory.DaysOfWeek[time.Monday] || !factory.DaysOfWeek[time.Wednesday] {
		t.Fatalf("expected daysOfWeek to include monday and wednesday, got %v", factory.DaysOfWeek)
	}
}

func TestRunDefaultIPAPIEndpointFreeAccount(t *testing.T) {
	path := writeConfig(t, `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8002"
    backendIP: "127.0.0.1"
    backendPort: "9002"
    allowedCountries: ["US"]
`)

	capture := &startCapture{}
	err := run([]string{"-config", path}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: capture.start,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	cfg := capture.configs[0]
	factory, ok := cfg.HandlerFactory.(*server.HandlerFactory)
	if !ok {
		t.Fatalf("expected HandlerFactory to be *server.HandlerFactory, got %T", cfg.HandlerFactory)
	}

	ipCfg, ok := factory.IPApiClient.(*ipapi.GetCountryCodeConfig)
	if !ok {
		t.Fatalf("expected IPApiClient to be *ipapi.GetCountryCodeConfig, got %T", factory.IPApiClient)
	}
	realClient, ok := ipCfg.HTTPClient.(*ipapi.RealHTTPClient)
	if !ok {
		t.Fatalf("expected HTTPClient to be *ipapi.RealHTTPClient, got %T", ipCfg.HTTPClient)
	}
	if realClient.Endpoint != "http://ip-api.com/json/" {
		t.Fatalf("unexpected default endpoint: %q", realClient.Endpoint)
	}
}

func TestRunDefaultIPAPIEndpointProAccount(t *testing.T) {
	path := writeConfig(t, `apiKey: "abc"
servers:
  - listenIP: "127.0.0.1"
    listenPort: "8003"
    backendIP: "127.0.0.1"
    backendPort: "9003"
    allowedCountries: ["US"]
`)

	capture := &startCapture{}
	err := run([]string{"-config", path}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: capture.start,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	cfg := capture.configs[0]
	factory, ok := cfg.HandlerFactory.(*server.HandlerFactory)
	if !ok {
		t.Fatalf("expected HandlerFactory to be *server.HandlerFactory, got %T", cfg.HandlerFactory)
	}

	ipCfg, ok := factory.IPApiClient.(*ipapi.GetCountryCodeConfig)
	if !ok {
		t.Fatalf("expected IPApiClient to be *ipapi.GetCountryCodeConfig, got %T", factory.IPApiClient)
	}
	realClient, ok := ipCfg.HTTPClient.(*ipapi.RealHTTPClient)
	if !ok {
		t.Fatalf("expected HTTPClient to be *ipapi.RealHTTPClient, got %T", ipCfg.HTTPClient)
	}
	if realClient.Endpoint != "https://pro.ip-api.com/json/" {
		t.Fatalf("unexpected default endpoint: %q", realClient.Endpoint)
	}
}

func TestRunRejectsHTTPSIPAPIWithoutAPIKey(t *testing.T) {
	path := writeConfig(t, `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8004"
    backendIP: "127.0.0.1"
    backendPort: "9004"
    allowedCountries: ["US"]
`)

	err := run([]string{"-config", path, "-ipapi", "https://ip-api.com/json/"}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: (&startCapture{}).start,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunRejectsIPAPIOverrideWithAPIKey(t *testing.T) {
	path := writeConfig(t, `apiKey: "abc"
servers:
  - listenIP: "127.0.0.1"
    listenPort: "8005"
    backendIP: "127.0.0.1"
    backendPort: "9005"
    allowedCountries: ["US"]
`)

	err := run([]string{"-config", path, "-ipapi", "http://127.0.0.1/json/"}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: (&startCapture{}).start,
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "no countries",
			content: `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8000"
    backendIP: "127.0.0.1"
    backendPort: "9000"
`,
			wantErr: "no countries specified",
		},
		{
			name: "start date after end date",
			content: `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8000"
    backendIP: "127.0.0.1"
    backendPort: "9000"
    allowedCountries: ["US"]
    startDate: "2024-01-02"
    endDate: "2024-01-01"
`,
			wantErr: "start date",
		},
		{
			name: "recv proxy protocol requires trusted proxies",
			content: `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8000"
    backendIP: "127.0.0.1"
    backendPort: "9000"
    allowedCountries: ["US"]
    recvProxyProtocol: true
`,
			wantErr: "trustedProxies",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			path := writeConfig(t, tc.content)
			err := run([]string{"-config", path}, runDeps{
				logger:      log.New(io.Discard, "", 0),
				flagOutput:  io.Discard,
				startServer: (&startCapture{}).start,
			})
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !bytes.Contains([]byte(err.Error()), []byte(tc.wantErr)) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestRunDefaultTimeouts(t *testing.T) {
	path := writeConfig(t, `servers:
  - listenIP: "127.0.0.1"
    listenPort: "8010"
    backendIP: "127.0.0.1"
    backendPort: "9010"
    allowedCountries: ["US"]
`)
	capture := &startCapture{}
	err := run([]string{"-config", path}, runDeps{
		logger:      log.New(io.Discard, "", 0),
		flagOutput:  io.Discard,
		startServer: capture.start,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if capture.calls != 1 {
		t.Fatalf("expected 1 startServer call, got %d", capture.calls)
	}
	factory, ok := capture.configs[0].HandlerFactory.(*server.HandlerFactory)
	if !ok {
		t.Fatalf("expected HandlerFactory to be *server.HandlerFactory, got %T", capture.configs[0].HandlerFactory)
	}
	if factory.IdleTimeout != 60*time.Second {
		t.Fatalf("expected idle timeout 60s, got %s", factory.IdleTimeout)
	}
	if factory.MaxConnLifetime != 2*time.Hour {
		t.Fatalf("expected max conn lifetime 2h, got %s", factory.MaxConnLifetime)
	}
}
