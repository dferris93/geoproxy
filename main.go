package main

import (
	"fmt"
	"geoproxy/common"
	"geoproxy/config"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"geoproxy/iptables"
	"geoproxy/server"
	"strings"
	"time"

	"context"
	"flag"
	"log"
	"sync"
)

func main() {
	configFile := flag.String("config", "geoproxy.yaml", "Path to the configuration file")
	continueOnError := flag.Bool("continue", false, "allow connections through on ipapi errors")
	blockIptables := flag.String("iptables", "", "add rejected IPs to the specified iptables chain")
	iptablesAction := flag.String("action", "DROP", "iptables action to take on blocked IPs. Default is DROP.")
	ipapiEndpoint := flag.String("ipapi", "http://ip-api.com/json/", "ipapi endpoint. If you have an API key, change this to https://pro.ip-api.com/json/")
	lruSize := flag.Int("lru", 10000, "size of the IP address LRU cache")
	flag.Parse()

	config, err := config.ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to read configuration file: %v", err)
	}

	log.Printf("Starting GeoProxy\n")
	log.Printf("Configuration file: %s\n", *configFile)
	log.Printf("Continue on error: %v\n", *continueOnError)
	log.Printf("Iptables chain: %s\n", *blockIptables)
	log.Printf("Iptables action: %s\n", *iptablesAction)
	log.Printf("IPAPI endpoint: %s\n", *ipapiEndpoint)
	log.Printf("LRU max cache size: %d\n", *lruSize)

	for _, c := range config.Servers {
		log.Print("----------")
		log.Printf("Server %s:%s\n", c.ListenIP, c.ListenPort)
		log.Printf("Backend %s:%s\n", c.BackendIP, c.BackendPort)
		log.Printf("Allowed countries: %v\n", c.AllowedCountries)
		log.Printf("Allowed regions: %v\n", c.AllowedRegions)
		log.Printf("Always allowed: %v\n", c.AlwaysAllowed)
		log.Printf("Always denied: %v\n", c.AlwaysDenied)
		log.Printf("Denied countries: %v\n", c.DeniedCountries)
		log.Printf("Denied regions: %v\n", c.DeniedRegions)
		log.Printf("RecvProxyProtocol: %v\n", c.RecvProxyProtocol)
		log.Printf("SendProxyProtocol: %v\n", c.SendProxyProtocol)
		log.Printf("ProxyProtocolVersion: %d\n", c.ProxyProtocolVersion)
		log.Printf("Days of week: %v\n", c.DaysOfWeek)
		log.Printf("Start date: %s\n", c.StartDate)
		log.Printf("End date: %s\n", c.EndDate)
		log.Printf("Start time: %s\n", c.StartTime)
		log.Printf("End time: %s\n", c.EndTime)

		if len(c.AllowedCountries) == 0 && len(c.DeniedCountries) == 0 {
			log.Fatalf("no countries specified for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if (c.StartDate != "" && c.EndDate == "") || (c.StartDate == "" && c.EndDate != "") {
			log.Fatalf("both startDate and endDate must be set for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if len(c.DaysOfWeek) > 0 && (c.StartDate != "" || c.EndDate != "") {
			log.Fatalf("daysOfWeek cannot be combined with startDate/endDate for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if c.StartDate != "" && c.EndDate != "" {
			startDate, err := time.ParseInLocation("2006-01-02", c.StartDate, time.Local)
			if err != nil {
				log.Fatalf("failed to parse start date %s: %v", c.StartDate, err)
			}
			endDate, err := time.ParseInLocation("2006-01-02", c.EndDate, time.Local)
			if err != nil {
				log.Fatalf("failed to parse end date %s: %v", c.EndDate, err)
			}
			if startDate.After(endDate) {
				log.Fatalf("start date %s is after end date %s", c.StartDate, c.EndDate)
			}
		}
		if (c.StartTime != "" && c.EndTime == "") || (c.StartTime == "" && c.EndTime != "") {
			log.Fatalf("both startTime and endTime must be set for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if c.StartTime != "" && c.EndTime != "" {
			_, err := time.ParseInLocation("15:04", c.StartTime, time.Local)
			if err != nil {
				log.Fatalf("failed to parse start time %s: %v", c.StartTime, err)
			}
			_, err = time.ParseInLocation("15:04", c.EndTime, time.Local)
			if err != nil {
				log.Fatalf("failed to parse end time %s: %v", c.EndTime, err)
			}
		}
		if len(c.DaysOfWeek) > 0 {
			_, err = parseDaysOfWeek(c.DaysOfWeek)
			if err != nil {
				log.Fatalf("failed to parse days of week %v: %v", c.DaysOfWeek, err)
			}
		}
	}

	var blockips chan string
	if *blockIptables != "" {
		blockips = make(chan string)
		iptablesConfig := iptables.IpTables{
			Chain:    *blockIptables,
			Action:   *iptablesAction,
			Runner:   &iptables.RealRunner{},
			CheckIPs: &common.CheckIPs{},
		}
		go iptablesConfig.BlockIPs(blockips, context.Background())
	}

	m := &sync.Mutex{}
	go ipapi.LRUCachedReplies(m, *lruSize)

	wg := sync.WaitGroup{}
	for _, c := range config.Servers {
		wg.Add(1)
		log.Printf("proxy server listening on %s:%s countries: %v regions: %v always allowed: %v always denied: %v",
			c.ListenIP,
			c.ListenPort,
			c.AllowedCountries,
			c.AllowedRegions,
			c.AlwaysAllowed,
			c.AlwaysDenied)

		var startTime time.Time
		var endTime time.Time
		var startDate time.Time
		var endDate time.Time
		if c.StartDate != "" && c.EndDate != "" {
			startDate, err = time.ParseInLocation("2006-01-02", c.StartDate, time.Local)
			if err != nil {
				log.Fatalf("failed to parse start date %s: %v", c.StartDate, err)
			}
			endDate, err = time.ParseInLocation("2006-01-02", c.EndDate, time.Local)
			if err != nil {
				log.Fatalf("failed to parse end date %s: %v", c.EndDate, err)
			}
		}
		if c.StartTime != "" && c.EndTime != "" {
			startTime, err = time.ParseInLocation("15:04", c.StartTime, time.Local)
			if err != nil {
				log.Fatalf("failed to parse start time %s: %v", c.StartTime, err)
			}
			endTime, err = time.ParseInLocation("15:04", c.EndTime, time.Local)
			if err != nil {
				log.Fatalf("failed to parse end time %s: %v", c.EndTime, err)
			}
		}
		daysOfWeek, err := parseDaysOfWeek(c.DaysOfWeek)
		if err != nil {
			log.Fatalf("failed to parse days of week %v: %v", c.DaysOfWeek, err)
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
			HandlerFactory: &server.HandlerFactory{
				IPApiClient: &ipapi.GetCountryCodeConfig{
					HTTPClient: &ipapi.RealHTTPClient{
						Endpoint: *ipapiEndpoint,
						APIKey:   config.APIKey,
					},
				},
				AllowedCountries: common.MakeSet(c.AllowedCountries),
				AllowedRegions:   common.MakeSet(c.AllowedRegions),
				DeniedCountries:  common.MakeSet(c.DeniedCountries),
				DeniedRegions:    common.MakeSet(c.DeniedRegions),
				AlwaysAllowed:    c.AlwaysAllowed,
				AlwaysDenied:     c.AlwaysDenied,
				ContinueOnError:  *continueOnError,
				CheckIps:         &common.CheckIPs{},
				Mutex:            m,
				TransferFunc:     handler.TransferData,
				IptablesBlock:    *iptablesAction != "",
				BlockIPs:         blockips,
				BackendIP:        c.BackendIP,
				BackendPort:      c.BackendPort,
				StartTime:        startTime,
				EndTime:          endTime,
				StartDate:        startDate,
				EndDate:          endDate,
				DaysOfWeek:       daysOfWeek,
			},
		}
		go s.StartServer(&wg, context.Background())
	}
	wg.Wait()
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
