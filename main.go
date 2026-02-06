package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/golang-lru/v2"

	"geoproxy/common"
	"geoproxy/config"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"geoproxy/server"
)

func main() {
	if err := run(os.Args[1:], runDeps{
		logger:     log.Default(),
		flagOutput: os.Stderr,
		startServer: func(s server.ServerConfig, wg *sync.WaitGroup, ctx context.Context) {
			go s.StartServer(wg, ctx)
		},
	}); err != nil {
		log.Fatalf("%v", err)
	}
}

type runDeps struct {
	logger      *log.Logger
	flagOutput  io.Writer
	startServer func(server.ServerConfig, *sync.WaitGroup, context.Context)
}

func run(args []string, deps runDeps) error {
	if deps.logger == nil {
		deps.logger = log.Default()
	}
	if deps.flagOutput == nil {
		deps.flagOutput = os.Stderr
	}
	if deps.startServer == nil {
		deps.startServer = func(s server.ServerConfig, wg *sync.WaitGroup, ctx context.Context) {
			go s.StartServer(wg, ctx)
		}
	}

	fs := flag.NewFlagSet("geoproxy", flag.ContinueOnError)
	fs.SetOutput(deps.flagOutput)
	configFile := fs.String("config", "geoproxy.yaml", "Path to the configuration file")
	ipapiEndpointFlag := fs.String("ipapi", "", "ipapi endpoint override (free accounts only). Defaults to http://ip-api.com/json/ when apiKey is empty. When apiKey is set, the endpoint is forced to https://pro.ip-api.com/json/ and cannot be overridden")
	ipapiTimeout := fs.Duration("ipapi-timeout", 5*time.Second, "timeout for ipapi HTTP requests (e.g. 5s)")
	ipapiMaxBytes := fs.Int64("ipapi-max-bytes", 1<<20, "maximum bytes to read from ipapi responses (default 1MiB)")
	backendDialTimeout := fs.Duration("backend-dial-timeout", 5*time.Second, "timeout for backend TCP dials (e.g. 5s)")
	idleTimeout := fs.Duration("idle-timeout", 0, "idle timeout for proxied connections (0 disables)")
	maxConnLifetime := fs.Duration("max-conn-lifetime", 24*time.Hour, "maximum lifetime for a proxied connection (0 disables; e.g. 24h)")
	maxConns := fs.Int("max-conns", 1024, "maximum concurrent client connections per server (0 disables)")
	proxyProtoTimeout := fs.Duration("proxyproto-timeout", 1*time.Second, "timeout for receiving HAProxy PROXY protocol headers from trusted proxies (e.g. 1s)")
	lruSize := fs.Int("lru", 10000, "size of the IP address LRU cache")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *maxConnLifetime < 0 {
		return fmt.Errorf("-max-conn-lifetime must be >= 0")
	}
	if *proxyProtoTimeout <= 0 {
		return fmt.Errorf("-proxyproto-timeout must be > 0")
	}

	cfg, err := config.ReadConfig(*configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	var ipapiEndpoint string
	if cfg.APIKey != "" {
		// Pro accounts must use HTTPS and we don't allow overriding it.
		if strings.TrimSpace(*ipapiEndpointFlag) != "" {
			return fmt.Errorf("-ipapi cannot be used when apiKey is set; endpoint is forced to https://pro.ip-api.com/json/")
		}
		ipapiEndpoint = "https://pro.ip-api.com/json/"
	} else {
		ipapiEndpoint = strings.TrimSpace(*ipapiEndpointFlag)
		if ipapiEndpoint == "" {
			ipapiEndpoint = "http://ip-api.com/json/"
		}
		if err := validateFreeIPAPIEndpoint(ipapiEndpoint); err != nil {
			return err
		}
	}

	deps.logger.Printf("Starting GeoProxy\n")
	deps.logger.Printf("Configuration file: %s\n", *configFile)
	deps.logger.Printf("IPAPI endpoint: %s\n", ipapiEndpoint)
	deps.logger.Printf("IPAPI timeout: %s\n", ipapiTimeout.String())
	deps.logger.Printf("IPAPI max bytes: %d\n", *ipapiMaxBytes)
	deps.logger.Printf("Backend dial timeout: %s\n", backendDialTimeout.String())
	deps.logger.Printf("Idle timeout: %s\n", idleTimeout.String())
	deps.logger.Printf("Max conn lifetime: %s\n", maxConnLifetime.String())
	deps.logger.Printf("Max conns: %d\n", *maxConns)
	deps.logger.Printf("Proxy protocol timeout: %s\n", proxyProtoTimeout.String())
	deps.logger.Printf("LRU cache size: %d\n", *lruSize)

	cache, err := lru.New[string, ipapi.Reply](*lruSize)
	if err != nil {
		return fmt.Errorf("failed to initialize IP cache: %v", err)
	}
	ipapi.IPCache = cache

	for _, c := range cfg.Servers {
		deps.logger.Print("----------")
		deps.logger.Printf("Server %s:%s\n", c.ListenIP, c.ListenPort)
		deps.logger.Printf("Backend %s:%s\n", c.BackendIP, c.BackendPort)
		deps.logger.Printf("Allowed countries: %v\n", c.AllowedCountries)
		deps.logger.Printf("Allowed regions: %v\n", c.AllowedRegions)
		deps.logger.Printf("Always allowed: %v\n", c.AlwaysAllowed)
		deps.logger.Printf("Always denied: %v\n", c.AlwaysDenied)
		deps.logger.Printf("Denied countries: %v\n", c.DeniedCountries)
		deps.logger.Printf("Denied regions: %v\n", c.DeniedRegions)
		deps.logger.Printf("RecvProxyProtocol: %v\n", c.RecvProxyProtocol)
		deps.logger.Printf("SendProxyProtocol: %v\n", c.SendProxyProtocol)
		deps.logger.Printf("ProxyProtocolVersion: %d\n", c.ProxyProtocolVersion)
		deps.logger.Printf("TrustedProxies: %v\n", c.TrustedProxies)
		deps.logger.Printf("Days of week: %v\n", c.DaysOfWeek)
		deps.logger.Printf("Start date: %s\n", c.StartDate)
		deps.logger.Printf("End date: %s\n", c.EndDate)
		deps.logger.Printf("Start time: %s\n", c.StartTime)
		deps.logger.Printf("End time: %s\n", c.EndTime)

		if len(c.AllowedCountries) == 0 && len(c.DeniedCountries) == 0 {
			return fmt.Errorf("no countries specified for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if c.SendProxyProtocol {
			if c.ProxyProtocolVersion != 1 && c.ProxyProtocolVersion != 2 {
				return fmt.Errorf("invalid proxyProtocolVersion %d for server %s:%s (expected 1 or 2)", c.ProxyProtocolVersion, c.ListenIP, c.ListenPort)
			}
		}
		if (c.StartDate != "" && c.EndDate == "") || (c.StartDate == "" && c.EndDate != "") {
			return fmt.Errorf("both startDate and endDate must be set for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if len(c.DaysOfWeek) > 0 && (c.StartDate != "" || c.EndDate != "") {
			return fmt.Errorf("daysOfWeek cannot be combined with startDate/endDate for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if c.StartDate != "" && c.EndDate != "" {
			startDate, err := time.ParseInLocation("2006-01-02", c.StartDate, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse start date %s: %v", c.StartDate, err)
			}
			endDate, err := time.ParseInLocation("2006-01-02", c.EndDate, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse end date %s: %v", c.EndDate, err)
			}
			if startDate.After(endDate) {
				return fmt.Errorf("start date %s is after end date %s", c.StartDate, c.EndDate)
			}
		}
		if (c.StartTime != "" && c.EndTime == "") || (c.StartTime == "" && c.EndTime != "") {
			return fmt.Errorf("both startTime and endTime must be set for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if c.StartTime != "" && c.EndTime != "" {
			_, err := time.ParseInLocation("15:04", c.StartTime, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse start time %s: %v", c.StartTime, err)
			}
			_, err = time.ParseInLocation("15:04", c.EndTime, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse end time %s: %v", c.EndTime, err)
			}
		}
		if len(c.DaysOfWeek) > 0 {
			_, err = parseDaysOfWeek(c.DaysOfWeek)
			if err != nil {
				return fmt.Errorf("failed to parse days of week %v: %v", c.DaysOfWeek, err)
			}
		}
		if c.RecvProxyProtocol && len(c.TrustedProxies) == 0 {
			return fmt.Errorf("recvProxyProtocol is true but trustedProxies is empty for server %s:%s; configure trustedProxies to avoid PROXY protocol spoofing", c.ListenIP, c.ListenPort)
		}
	}

	wg := sync.WaitGroup{}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for _, c := range cfg.Servers {
		wg.Add(1)
		deps.logger.Printf("proxy server listening on %s:%s countries: %v regions: %v always allowed: %v always denied: %v",
			c.ListenIP,
			c.ListenPort,
			c.AllowedCountries,
			c.AllowedRegions,
			c.AlwaysAllowed,
			c.AlwaysDenied)

		trustedProxies := c.TrustedProxies
		if !c.RecvProxyProtocol {
			if len(trustedProxies) > 0 {
				deps.logger.Printf("trustedProxies ignored because recvProxyProtocol is false on %s:%s", c.ListenIP, c.ListenPort)
			}
			trustedProxies = nil
		}

		var startTime time.Time
		var endTime time.Time
		var startDate time.Time
		var endDate time.Time
		if c.StartDate != "" && c.EndDate != "" {
			startDate, err = time.ParseInLocation("2006-01-02", c.StartDate, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse start date %s: %v", c.StartDate, err)
			}
			endDate, err = time.ParseInLocation("2006-01-02", c.EndDate, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse end date %s: %v", c.EndDate, err)
			}
		}
		if c.StartTime != "" && c.EndTime != "" {
			startTime, err = time.ParseInLocation("15:04", c.StartTime, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse start time %s: %v", c.StartTime, err)
			}
			endTime, err = time.ParseInLocation("15:04", c.EndTime, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse end time %s: %v", c.EndTime, err)
			}
		}
		daysOfWeek, err := parseDaysOfWeek(c.DaysOfWeek)
		if err != nil {
			return fmt.Errorf("failed to parse days of week %v: %v", c.DaysOfWeek, err)
		}
		s := server.ServerConfig{
			ListenIP:          c.ListenIP,
			ListenPort:        c.ListenPort,
			BackendIP:         c.BackendIP,
			BackendPort:       c.BackendPort,
			NetListener:       &server.RealNetListener{},
			RecvProxyProtocol: c.RecvProxyProtocol,
			TrustedProxies:    trustedProxies,
			MaxConns:          *maxConns,
			ProxyProtoTimeout: *proxyProtoTimeout,
			HandlerFactory: &server.HandlerFactory{
				BackendDialer: &server.RealDialer{Timeout: *backendDialTimeout},
				IPApiClient: &ipapi.GetCountryCodeConfig{
					HTTPClient: &ipapi.RealHTTPClient{
						Endpoint: ipapiEndpoint,
						APIKey:   cfg.APIKey,
						Timeout:  *ipapiTimeout,
					},
					Cache:            ipapi.IPCache,
					MaxResponseBytes: *ipapiMaxBytes,
				},
				AllowedCountries:     common.MakeSet(c.AllowedCountries),
				AllowedRegions:       common.MakeSet(c.AllowedRegions),
				DeniedCountries:      common.MakeSet(c.DeniedCountries),
				DeniedRegions:        common.MakeSet(c.DeniedRegions),
				AlwaysAllowed:        c.AlwaysAllowed,
				AlwaysDenied:         c.AlwaysDenied,
				CheckIps:             &common.CheckIPs{},
				TransferFunc:         handler.TransferData,
				BackendIP:            c.BackendIP,
				BackendPort:          c.BackendPort,
				SendProxyProtocol:    c.SendProxyProtocol,
				ProxyProtocolVersion: c.ProxyProtocolVersion,
				MaxConnLifetime:      *maxConnLifetime,
				StartTime:            startTime,
				EndTime:              endTime,
				StartDate:            startDate,
				EndDate:              endDate,
				DaysOfWeek:           daysOfWeek,
				IdleTimeout:          *idleTimeout,
			},
		}
		deps.startServer(s, &wg, ctx)
	}
	wg.Wait()
	return nil
}

func validateFreeIPAPIEndpoint(endpoint string) error {
	// Operator-provided override should still be a sane HTTP URL.
	// (Free ip-api is HTTP-only.)
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid ipapi endpoint %q: %v", endpoint, err)
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme == "https" {
		return fmt.Errorf("ipapi endpoint %q uses https but apiKey is empty; free ip-api does not support SSL", endpoint)
	}
	if scheme != "http" && scheme != "" {
		return fmt.Errorf("invalid ipapi endpoint %q: scheme must be http", endpoint)
	}
	if u.User != nil {
		return fmt.Errorf("invalid ipapi endpoint %q: userinfo not allowed", endpoint)
	}
	if strings.TrimSpace(u.Host) == "" {
		// url.Parse treats "http://..." as having Host; bare hosts can end up in Path.
		return fmt.Errorf("invalid ipapi endpoint %q: missing host", endpoint)
	}
	return nil
}

func parseDaysOfWeek(days []string) (map[time.Weekday]bool, error) {
	if len(days) == 0 {
		return map[time.Weekday]bool{}, nil
	}
	parsed := make(map[time.Weekday]bool, len(days))
	for _, day := range days {
		weekday, err := parseWeekday(day)
		if err != nil {
			return nil, err
		}
		parsed[weekday] = true
	}
	return parsed, nil
}

func parseWeekday(day string) (time.Weekday, error) {
	normalized := strings.ToLower(strings.TrimSpace(day))
	switch normalized {
	case "sun", "sunday":
		return time.Sunday, nil
	case "mon", "monday":
		return time.Monday, nil
	case "tue", "tues", "tuesday":
		return time.Tuesday, nil
	case "wed", "wednesday":
		return time.Wednesday, nil
	case "thu", "thurs", "thursday":
		return time.Thursday, nil
	case "fri", "friday":
		return time.Friday, nil
	case "sat", "saturday":
		return time.Saturday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid weekday: %s", day)
	}
}
