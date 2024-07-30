package ipapi

import (
	"fmt"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type RealHTTPClient struct{
	Endpoint string
	APIKey string
}

func (r *RealHTTPClient) Get(ip string) (*http.Response, error) {
	if r.APIKey != "" {
		return http.Get(fmt.Sprintf("%s/%s?key=%s&fields=countryCode,region,status", r.Endpoint, ip, r.APIKey))
	}
	return http.Get(r.Endpoint + ip)
}