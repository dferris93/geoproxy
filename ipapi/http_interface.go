package ipapi

import (
	"net/http"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type RealHTTPClient struct{}

func (r *RealHTTPClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}