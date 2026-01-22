package ipapi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/patrickmn/go-cache"
)

var IPCache *cache.Cache

type IPAPI interface {
	GetCountryCode(ip string) (string, string, string, error)
}

type Reply struct {
	CountryCode string
	Region      string
}
type GetCountryCodeConfig struct {
	HTTPClient HTTPClient
	Cache      *cache.Cache
}

func (g *GetCountryCodeConfig) GetCountryCode(ip string) (string, string, string, error) {
	if IPCache != nil {
		if cachedReply, found := IPCache.Get(ip); found {
			reply := cachedReply.(Reply)
			return reply.CountryCode, reply.Region, "cached", nil
		}
	}
	ipAPIConfig := &IPAPIConfig{HTTPClient: g.HTTPClient}
	countryCode, region, err := ipAPIConfig.getIpAPI(ip)
	if err != nil {
		return "", "", "-", err
	}
	if IPCache != nil {
		IPCache.Set(ip, Reply{CountryCode: countryCode, Region: region}, cache.DefaultExpiration)
	}
	return countryCode, region, "-", nil
}

type IPAPIConfig struct {
	HTTPClient HTTPClient
}

func (i *IPAPIConfig) getIpAPI(ip string) (string, string, error) {
	// PathEscape the IP to prevent path traversal (SSRF)
	escapedIP := url.PathEscape(ip)
	resp, err := i.HTTPClient.Get(escapedIP)
	if err != nil {
		return "", "", fmt.Errorf("failed to get country code: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		CountryCode string `json:"countryCode"`
		Region      string `json:"region"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if data.Status != "success" {
		return "--", "--", fmt.Errorf("failed to get country code for ip: %s", ip)
	}
	return data.CountryCode, data.Region, nil
}

// LRUCachedReplies is a no-op because the vulnerable caching has been removed.
func LRUCachedReplies(lruSize int) {
	// Caching removed due to security vulnerabilities.
}
