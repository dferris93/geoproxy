package common

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type CheckIP interface {
	CheckIPType(ip string) (int, error)
	CheckSubnets(subnets []string, clientAddr string) bool
}

type CheckIPs struct{}

func (c *CheckIPs) CheckIPType(ip string) (int, error) {
	ipaddr := net.ParseIP(ip)

	if ipaddr == nil {
		return 0, fmt.Errorf("failed to parse ip")
	}

	if ipaddr.To4() != nil {
		return 4, nil
	}

	return 6, nil
}


func (c *CheckIPs) CheckSubnets(subnets []string, clientAddr string) bool {
	for _, ip := range subnets {
		if !strings.Contains(ip, "/") {
			ipType, err := c.CheckIPType(ip)
			if err != nil {
				log.Printf("Failed to check IP type: %v", err)
				continue
			}
			if ipType == 4 {
				ip += "/32"
			} else {
				ip += "/128"
			}
		}
		_, subnet, err := net.ParseCIDR(ip)
		if err != nil {
			log.Printf("Failed to parse CIDR: %v", err)
			continue
		}	
		if subnet.Contains(net.ParseIP(clientAddr)) {
			return true	
		}
	}
	return false
}

func MakeSet(s []string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range s {
		set[item] = true
	}
	return set
}

func SubtractSet(a, b map[string]bool) map[string]bool {
	for key := range b {
		delete(a, key)
	}
	return a
}

func CheckTime(startTime time.Time, endTime time.Time, now time.Time) (bool, error) {
	current := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)

	if startTime.Before(endTime) {
		return current.After(startTime) && current.Before(endTime), nil
	} else {
		return current.After(startTime) || current.Before(endTime), nil
	}
}