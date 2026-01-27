package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"geoproxy/common"
	"geoproxy/config"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"geoproxy/server"
	"strings"
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
	continueOnError := fs.Bool("continue", false, "allow connections through on ipapi errors")
	ipapiEndpoint := fs.String("ipapi", "http://ip-api.com/json/", "ipapi endpoint. If you have an API key, change this to https://pro.ip-api.com/json/")
	cacheExpiration := fs.Duration("cache-expiration", 1*time.Hour, "Default expiration for cache entries")
	cacheCleanup := fs.Duration("cache-cleanup", 10*time.Minute, "Interval to clean up expired cache entries")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.ReadConfig(*configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	deps.logger.Printf("Starting GeoProxy\n")
	deps.logger.Printf("Configuration file: %s\n", *configFile)
	deps.logger.Printf("Continue on error: %v\n", *continueOnError)
	deps.logger.Printf("IPAPI endpoint: %s\n", *ipapiEndpoint)
	deps.logger.Printf("Cache expiration: %s\n", *cacheExpiration)
	deps.logger.Printf("Cache cleanup interval: %s\n", *cacheCleanup)

	ipapi.IPCache = cache.New(*cacheExpiration, *cacheCleanup)

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
	}

	wg := sync.WaitGroup{}
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
			ListenIP:             c.ListenIP,
			ListenPort:           c.ListenPort,
			BackendIP:            c.BackendIP,
			BackendPort:          c.BackendPort,
			NetListener:          &server.RealNetListener{},
			Dialer:               &server.RealDialer{},
			RecvProxyProtocol:    c.RecvProxyProtocol,
			SendProxyProtocol:    c.SendProxyProtocol,
			ProxyProtocolVersion: c.ProxyProtocolVersion,
			TrustedProxies:       trustedProxies,
			HandlerFactory: &server.HandlerFactory{
				IPApiClient: &ipapi.GetCountryCodeConfig{
					HTTPClient: &ipapi.RealHTTPClient{
						Endpoint: *ipapiEndpoint,
						APIKey:   cfg.APIKey,
					},
					Cache: ipapi.IPCache,
				},
				AllowedCountries: common.MakeSet(c.AllowedCountries),
				AllowedRegions:   common.MakeSet(c.AllowedRegions),
				DeniedCountries:  common.MakeSet(c.DeniedCountries),
				DeniedRegions:    common.MakeSet(c.DeniedRegions),
				AlwaysAllowed:    c.AlwaysAllowed,
				AlwaysDenied:     c.AlwaysDenied,
				ContinueOnError:  *continueOnError,
				CheckIps:         &common.CheckIPs{},
				TransferFunc:     handler.TransferData,
				BackendIP:        c.BackendIP,
				BackendPort:      c.BackendPort,
				StartTime:        startTime,
				EndTime:          endTime,
				StartDate:        startDate,
				EndDate:          endDate,
				DaysOfWeek:       daysOfWeek,
			},
		}
		deps.startServer(s, &wg, context.Background())
	}
	wg.Wait()
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
