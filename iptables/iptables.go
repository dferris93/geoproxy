package iptables

import (
	"context"
	"geoproxy/common"
	"log"
)

type IpTables struct {
	Chain string
	Action string
	Runner Runner
	CheckIPs common.CheckIP
}

func (i *IpTables) BlockIPs(blockIPs chan string, ctx context.Context) error {
	blockedAddresses := make(map[string]bool)	
	for {
		select {
		case <-ctx.Done():
			return nil
		case ip, ok := <-blockIPs:
			if !ok {
				return nil
			}
			if blockedAddresses[ip] {
				continue
			}
			ipType, err := i.CheckIPs.CheckIPType(ip)
			if err != nil {
				log.Printf("Failed to check IP type: %v", err)	
				continue
			}
			if ipType == 4 {
				_, err = i.Runner.RunCommand("iptables", "-A", i.Chain, "-s", ip, "-j", i.Action)
			} else {
				_, err = i.Runner.RunCommand("ip6tables", "-A", i.Chain, "-s", ip, "-j", i.Action)
			}
			if err != nil {
				log.Printf("Failed to block ip: %v", err)
			}	
			log.Printf("Blocked ip: %s", ip)
			blockedAddresses[ip] = true
		}
	}
}
