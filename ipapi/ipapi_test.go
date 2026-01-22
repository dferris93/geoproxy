package ipapi

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

type mockHTTPClient struct {
	getFunc func(url string) (*http.Response, error)
	calls   int
	lastURL string
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	m.calls++
	m.lastURL = url
	return m.getFunc(url)
}

func responseWithBody(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestIPAPIGetIpAPISuccess(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"US","region":"CA","status":"success"}`), nil
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	country, region, err := cfg.getIpAPI("1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, 1, client.calls)
	assert.Equal(t, "1.2.3.4", client.lastURL)
}

func TestIPAPIGetIpAPIDecodeError(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`not-json`), nil
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	_, _, err := cfg.getIpAPI("1.2.3.4")

	assert.Error(t, err)
}

func TestIPAPIGetIpAPIStatusError(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"--","region":"--","status":"fail"}`), nil
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	country, region, err := cfg.getIpAPI("1.2.3.4")

	assert.Error(t, err)
	assert.Equal(t, "--", country)
	assert.Equal(t, "--", region)
}

func TestIPAPIGetIpAPIClientError(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	_, _, err := cfg.getIpAPI("1.2.3.4")

	assert.Error(t, err)
}

func TestGetCountryCodeCacheHit(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := cache.New(5*time.Minute, 10*time.Minute)
	IPCache = tempCache // Set the global cache for this test
	
	tempCache.Set("1.2.3.4", Reply{CountryCode: "US", Region: "CA"}, cache.DefaultExpiration)
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode("1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, "cached", cacheMarker)
	assert.Equal(t, 0, client.calls)
}

func TestGetCountryCodeCacheMiss(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := cache.New(5*time.Minute, 10*time.Minute)
	IPCache = tempCache // Set the global cache for this test

	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"US","region":"CA","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode("1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)
	_, found := tempCache.Get("1.2.3.4")
	assert.True(t, found)
}

func TestGetCountryCodeCacheExpired(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := cache.New(5*time.Minute, 10*time.Minute)
	IPCache = tempCache // Set the global cache for this test

	tempCache.Set("1.2.3.4", Reply{CountryCode: "US", Region: "CA"}, 5*time.Millisecond) // Set a short-lived item
	time.Sleep(10 * time.Millisecond)                                                   // Ensure it expires before lookup
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode("1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "DE", country)
	assert.Equal(t, "BE", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)
	cached, found := tempCache.Get("1.2.3.4")
	assert.True(t, found)
	assert.Equal(t, "DE", cached.(Reply).CountryCode)
}

func TestGetCountryCodeCacheMissError(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := cache.New(5*time.Minute, 10*time.Minute)
	IPCache = tempCache // Set the global cache for this test

	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	_, _, cacheMarker, err := cfg.GetCountryCode("1.2.3.4")

	assert.Error(t, err)
	assert.Equal(t, "-", cacheMarker)
	_, found := tempCache.Get("1.2.3.4")
	assert.False(t, found) // Should not cache on error
}
