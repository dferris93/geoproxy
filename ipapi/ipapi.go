package ipapi

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)	

var (
	CachedReplies = map[string]Reply{}
)

type IPAPI interface{
	GetCountryCode(ip string, m *sync.Mutex) (string, string, string, error)
}

type Reply struct {
	TimeStamp   time.Time
	CountryCode string
	Region      string
}
type GetCountryCodeConfig struct {
	HTTPClient HTTPClient
}

func (g *GetCountryCodeConfig) GetCountryCode(ip string, m *sync.Mutex) (string, string, string, error) {
	var countryCode, region string
	var err error
	ipAPIConfig := &IPAPIConfig{HTTPClient: g.HTTPClient}
	if reply, ok := CachedReplies[ip]; ok {
		if time.Since(reply.TimeStamp) >= 24*time.Hour {
			countryCode, region, err = ipAPIConfig.getIpAPI(ip)
			if err != nil {
				return "", "", "", err
			}
			m.Lock()
			defer m.Unlock()
			CachedReplies[ip] = Reply{time.Now(), countryCode, region}
			return countryCode, region, "-", nil
		} else {
			return reply.CountryCode, reply.Region, "*", nil
		}
	} else {
		countryCode, region, err := ipAPIConfig.getIpAPI(ip)
		if err != nil {
			return "", "", "-", err
		}
		m.Lock()
		defer m.Unlock()
		CachedReplies[ip] = Reply{time.Now(), countryCode, region}
		return countryCode, region, "-", nil
	}
}

type IPAPIConfig struct {
	HTTPClient HTTPClient
}

func (i* IPAPIConfig) getIpAPI(ip string) (string, string, error) {
	resp, err := i.HTTPClient.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "", "", fmt.Errorf("failed to get country code: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		CountryCode string `json:"countryCode"`
		Region      string `json:"region"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}
	return data.CountryCode, data.Region, nil
}
