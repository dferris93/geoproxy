package main

import (
	"geoproxy/common"
	"geoproxy/config"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"geoproxy/iptables"
	"geoproxy/server"
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

	var startTime time.Time
	var endTime time.Time

	for _, c := range config.Servers {
		if len(c.AllowedCountries) == 0 && len(c.DeniedCountries) == 0 {
			log.Fatalf("no countries specified for server %s:%s", c.ListenIP, c.ListenPort)
		}
		if c.StartTime != "" && c.EndTime != "" {
			startTime, err = time.Parse("15:04", c.StartTime)
			if err != nil {
				log.Fatalf("failed to parse start time %s: %v", c.StartTime, err)
			}
			endTime, err = time.Parse("15:04", c.EndTime)
			if err != nil {
				log.Fatalf("failed to parse end time %s: %v", c.EndTime, err)
			}
			if startTime.After(endTime) {
				log.Fatalf("start time %s is after end time %s", c.StartTime, c.EndTime)
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
			},
		}
		go s.StartServer(&wg, context.Background())
	}
	wg.Wait()
}
