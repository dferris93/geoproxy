package ipapi

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)	

var (
	CachedReplies = map[string]Reply{}
	LRUOrder []string
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
			go updateLRUOrder(ip, m)
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
		go updateLRUOrder(ip, m)
		return countryCode, region, "-", nil
	}
}

type IPAPIConfig struct {
	HTTPClient HTTPClient
}

func (i* IPAPIConfig) getIpAPI(ip string) (string, string, error) {
	resp, err := i.HTTPClient.Get(ip)
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

func updateLRUOrder(key string, m *sync.Mutex) {
	for i, v := range LRUOrder {
		if v == key {
			m.Lock()
			defer m.Unlock()
			LRUOrder = append(LRUOrder[:i], LRUOrder[i+1:]...)
			break
		}
	}
	LRUOrder = append(LRUOrder, key)
}

func LRUCachedReplies(m *sync.Mutex, lruSize int) {
	for {
		time.Sleep(1 * time.Minute)
		lruEntries := len(CachedReplies)
		log.Printf("LRU cache size: %d\n", lruEntries)
		if lruEntries > lruSize {
			m.Lock()
			for len(CachedReplies) > lruSize {
				oldestKey := LRUOrder[0]
				delete(CachedReplies, oldestKey)
				LRUOrder = LRUOrder[1:]
			}
			m.Unlock()
		}
	}
}
