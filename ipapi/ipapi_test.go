package ipapi

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

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

func resetCache() {
	CachedReplies = map[string]Reply{}
	LRUOrder = []string{}
}

func TestIPAPIGetIpAPISuccess(t *testing.T) {
	resetCache()
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
	resetCache()
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
	resetCache()
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
	resetCache()
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
	resetCache()
	CachedReplies["1.2.3.4"] = Reply{
		TimeStamp:   time.Now().Add(-1 * time.Hour),
		CountryCode: "US",
		Region:      "CA",
	}
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}
	mutex := &sync.Mutex{}

	country, region, cacheMarker, err := cfg.GetCountryCode("1.2.3.4", mutex)

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, "*", cacheMarker)
	assert.Equal(t, 0, client.calls)
}

func TestGetCountryCodeCacheMiss(t *testing.T) {
	resetCache()
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"US","region":"CA","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}
	mutex := &sync.Mutex{}

	country, region, cacheMarker, err := cfg.GetCountryCode("1.2.3.4", mutex)

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)
	assert.Contains(t, CachedReplies, "1.2.3.4")
}

func TestGetCountryCodeCacheExpired(t *testing.T) {
	resetCache()
	CachedReplies["1.2.3.4"] = Reply{
		TimeStamp:   time.Now().Add(-25 * time.Hour),
		CountryCode: "US",
		Region:      "CA",
	}
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}
	mutex := &sync.Mutex{}

	country, region, cacheMarker, err := cfg.GetCountryCode("1.2.3.4", mutex)

	assert.NoError(t, err)
	assert.Equal(t, "DE", country)
	assert.Equal(t, "BE", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)
	assert.Equal(t, "DE", CachedReplies["1.2.3.4"].CountryCode)
}

func TestGetCountryCodeCacheMissError(t *testing.T) {
	resetCache()
	client := &mockHTTPClient{
		getFunc: func(url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}
	mutex := &sync.Mutex{}

	_, _, cacheMarker, err := cfg.GetCountryCode("1.2.3.4", mutex)

	assert.Error(t, err)
	assert.Equal(t, "-", cacheMarker)
}
