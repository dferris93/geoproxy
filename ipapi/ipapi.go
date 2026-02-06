package ipapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/hashicorp/golang-lru/v2"
)

var IPCache *lru.Cache[string, Reply]

type IPAPI interface {
	GetCountryCode(ctx context.Context, ip string) (string, string, string, error)
}

type Reply struct {
	CountryCode string
	Region      string
}
type GetCountryCodeConfig struct {
	HTTPClient       HTTPClient
	Cache            *lru.Cache[string, Reply]
	MaxResponseBytes int64
}

func (g *GetCountryCodeConfig) GetCountryCode(ctx context.Context, ip string) (string, string, string, error) {
	cache := g.Cache
	if cache == nil {
		cache = IPCache
	}
	if cache != nil {
		if reply, found := cache.Get(ip); found {
			return reply.CountryCode, reply.Region, "cached", nil
		}
	}
	ipAPIConfig := &IPAPIConfig{HTTPClient: g.HTTPClient, MaxResponseBytes: g.MaxResponseBytes}
	countryCode, region, err := ipAPIConfig.getIpAPI(ctx, ip)
	if err != nil {
		return "", "", "-", err
	}
	if cache != nil {
		cache.Add(ip, Reply{CountryCode: countryCode, Region: region})
	}
	return countryCode, region, "-", nil
}

type IPAPIConfig struct {
	HTTPClient       HTTPClient
	MaxResponseBytes int64
}

func (i *IPAPIConfig) getIpAPI(ctx context.Context, ip string) (string, string, error) {
	// PathEscape the IP to prevent path traversal (SSRF)
	escapedIP := url.PathEscape(ip)
	resp, err := i.HTTPClient.Get(ctx, escapedIP)
	if err != nil {
		return "", "", fmt.Errorf("failed to get country code: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("ipapi returned non-200 status: %d", resp.StatusCode)
	}

	limit := i.MaxResponseBytes
	if limit <= 0 {
		limit = 1 << 20 // 1MiB
	}
	limited := io.LimitReader(resp.Body, limit)

	var data struct {
		CountryCode string `json:"countryCode"`
		Region      string `json:"region"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(limited).Decode(&data); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	if data.Status != "success" {
		return "--", "--", fmt.Errorf("failed to get country code for ip: %s", ip)
	}
	return data.CountryCode, data.Region, nil
}
