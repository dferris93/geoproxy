package main

import (
	"geoproxy/common"
	"geoproxy/config"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"geoproxy/iptables"
	"geoproxy/server"

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
	flag.Parse()

	config, err := config.ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to read configuration file: %v", err)
	}

	for _, c := range config.Servers {
		if len(c.AllowedCountries) == 0 && len(c.DeniedCountries) == 0 {
			log.Fatalf("no countries specified for server %s:%s", c.ListenIP, c.ListenPort)
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
			ListenIP:    c.ListenIP,
			ListenPort:  c.ListenPort,
			BackendIP:   c.BackendIP,
			BackendPort: c.BackendPort,
			NetListener: &server.RealNetListener{},
			Dialer:      &server.RealDialer{},
			HandlerFactory: &server.HandlerFactory{
				IPApiClient: 	 &ipapi.GetCountryCodeConfig{HTTPClient: &ipapi.RealHTTPClient{}},
				AllowedCountries: common.MakeSet(c.AllowedCountries),
				AllowedRegions:   common.MakeSet(c.AllowedRegions),
				DeniedCountries:  common.MakeSet(c.DeniedCountries),
				DeniedRegions:    common.MakeSet(c.DeniedRegions),
				AlwaysAllowed:    c.AlwaysAllowed,
				AlwaysDenied:     c.AlwaysDenied,
				ContinueOnError:  *continueOnError,
				CheckIps:         &common.CheckIPs{},
				Mutex:            &sync.Mutex{},
				TransferFunc:     handler.TransferData,
				IptablesBlock:    *iptablesAction != "",
				BlockIPs: 	   blockips,
				BackendIP: 	  c.BackendIP,
				BackendPort:  c.BackendPort,
			},
		}
		go s.StartServer(&wg, context.Background())
	}
	wg.Wait()
}
