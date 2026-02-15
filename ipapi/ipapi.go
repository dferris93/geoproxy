package ipapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/hashicorp/golang-lru/v2"
)

var IPCache *lru.Cache[string, Reply]

type IPAPI interface {
	GetCountryCode(ctx context.Context, ip string) (string, string, string, error)
}

type Reply struct {
	CountryCode  string
	Region       string
	FailureUntil time.Time
	ExpiresAt    time.Time
}

const successCacheTTL = 24 * time.Hour

type GetCountryCodeConfig struct {
	HTTPClient       HTTPClient
	Cache            *lru.Cache[string, Reply]
	MaxResponseBytes int64
	FailureTTL       time.Duration
}

func (g *GetCountryCodeConfig) GetCountryCode(ctx context.Context, ip string) (string, string, string, error) {
	cache := g.Cache
	if cache == nil {
		cache = IPCache
	}
	if cache != nil {
		if reply, found := cache.Get(ip); found {
			if !reply.FailureUntil.IsZero() {
				if time.Now().Before(reply.FailureUntil) {
					return "", "", "cached-failure", fmt.Errorf("cached ipapi lookup failure for ip: %s", ip)
				}
				cache.Remove(ip)
			} else {
				// Successful cache entries are valid for 24 hours.
				if reply.ExpiresAt.IsZero() || time.Now().After(reply.ExpiresAt) {
					cache.Remove(ip)
				} else {
					return reply.CountryCode, reply.Region, "cached", nil
				}
			}
		}
	}
	ipAPIConfig := &IPAPIConfig{HTTPClient: g.HTTPClient, MaxResponseBytes: g.MaxResponseBytes}
	countryCode, region, err := ipAPIConfig.getIpAPI(ctx, ip)
	if err != nil {
		if cache != nil && g.FailureTTL > 0 {
			cache.Add(ip, Reply{FailureUntil: time.Now().Add(g.FailureTTL)})
		}
		return "", "", "-", err
	}
	if cache != nil {
		cache.Add(ip, Reply{
			CountryCode: countryCode,
			Region:      region,
			ExpiresAt:   time.Now().Add(successCacheTTL),
		})
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
