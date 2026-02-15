package ipapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient interface {
	// Get fetches ip-api data for the given (already path-escaped) IP string.
	Get(ctx context.Context, ip string) (*http.Response, error)
}

type RealHTTPClient struct {
	Endpoint string
	APIKey   string
	Timeout  time.Duration
	Client   *http.Client
}

func (r *RealHTTPClient) httpClient() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &http.Client{Timeout: timeout}
}

func (r *RealHTTPClient) buildURL(ip string) (string, error) {
	base, err := url.Parse(r.Endpoint)
	if err != nil {
		return "", err
	}

	// Join "<endpoint>/<ip>" with exactly one slash between.
	base.Path = strings.TrimRight(base.Path, "/") + "/" + ip

	q := base.Query()
	// Always request only the fields we actually use.
	q.Set("fields", "countryCode,region,status")
	base.RawQuery = q.Encode()

	return base.String(), nil
}

func (r *RealHTTPClient) Get(ctx context.Context, ip string) (*http.Response, error) {
	u, err := r.buildURL(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to build ipapi url: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	if r.APIKey != "" {
		req.Header.Set("X-API-Key", r.APIKey)
	}
	return r.httpClient().Do(req)
}
