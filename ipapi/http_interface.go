package ipapi

import (
	"net/http"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type RealHTTPClient struct{
	Endpoint string
}

func (r *RealHTTPClient) Get(ip string) (*http.Response, error) {
	return http.Get(r.Endpoint + ip)
}