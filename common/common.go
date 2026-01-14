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
	location := now.Location()
	current := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, location)
	start := time.Date(0, 1, 1, startTime.Hour(), startTime.Minute(), 0, 0, location)
	end := time.Date(0, 1, 1, endTime.Hour(), endTime.Minute(), 0, 0, location)

	if start.Before(end) {
		return (current.Equal(start) || current.After(start)) && (current.Equal(end) || current.Before(end)), nil
	} else {
		return current.Equal(start) || current.Equal(end) || current.After(start) || current.Before(end), nil
	}
}

func CheckDateRange(startDate time.Time, endDate time.Time, now time.Time) (bool, error) {
	location := now.Location()
	current := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, location)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, location)

	if start.After(end) {
		return false, fmt.Errorf("start date is after end date")
	}

	return (current.Equal(start) || current.After(start)) && (current.Equal(end) || current.Before(end)), nil
}
