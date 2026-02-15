package ipapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"github.com/stretchr/testify/assert"
)

type mockHTTPClient struct {
	getFunc func(ctx context.Context, url string) (*http.Response, error)
	calls   int
	lastURL string
}

func (m *mockHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	m.calls++
	m.lastURL = url
	return m.getFunc(ctx, url)
}

func responseWithBody(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func newTestCache(t *testing.T, size int) *lru.Cache[string, Reply] {
	t.Helper()
	cache, err := lru.New[string, Reply](size)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	return cache
}

func TestIPAPIGetIpAPISuccess(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"US","region":"CA","status":"success"}`), nil
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	country, region, err := cfg.getIpAPI(context.Background(), "1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, 1, client.calls)
	assert.Equal(t, "1.2.3.4", client.lastURL)
}

func TestIPAPIGetIpAPIDecodeError(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`not-json`), nil
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	_, _, err := cfg.getIpAPI(context.Background(), "1.2.3.4")

	assert.Error(t, err)
}

func TestIPAPIGetIpAPIStatusError(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"--","region":"--","status":"fail"}`), nil
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	country, region, err := cfg.getIpAPI(context.Background(), "1.2.3.4")

	assert.Error(t, err)
	assert.Equal(t, "--", country)
	assert.Equal(t, "--", region)
}

func TestIPAPIGetIpAPIClientError(t *testing.T) {
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &IPAPIConfig{HTTPClient: client}

	_, _, err := cfg.getIpAPI(context.Background(), "1.2.3.4")

	assert.Error(t, err)
}

func TestGetCountryCodeCacheHit(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := newTestCache(t, 16)
	IPCache = tempCache // Set the global cache for this test

	tempCache.Add("1.2.3.4", Reply{
		CountryCode: "US",
		Region:      "CA",
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode(context.Background(), "1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, "cached", cacheMarker)
	assert.Equal(t, 0, client.calls)
}

func TestGetCountryCodeCacheMiss(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := newTestCache(t, 16)
	IPCache = tempCache // Set the global cache for this test

	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"US","region":"CA","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode(context.Background(), "1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "US", country)
	assert.Equal(t, "CA", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)
	_, found := tempCache.Get("1.2.3.4")
	assert.True(t, found)
}

func TestGetCountryCodeCacheEvicted(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := newTestCache(t, 1)
	IPCache = tempCache // Set the global cache for this test

	tempCache.Add("1.2.3.4", Reply{
		CountryCode: "US",
		Region:      "CA",
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode(context.Background(), "5.6.7.8")

	assert.NoError(t, err)
	assert.Equal(t, "DE", country)
	assert.Equal(t, "BE", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)
	_, found := tempCache.Get("1.2.3.4")
	assert.False(t, found)
	cached, found := tempCache.Get("5.6.7.8")
	assert.True(t, found)
	assert.Equal(t, "DE", cached.CountryCode)
}

func TestGetCountryCodeExpiredSuccessCacheRefreshes(t *testing.T) {
	tempCache := newTestCache(t, 16)
	IPCache = tempCache

	tempCache.Add("1.2.3.4", Reply{
		CountryCode: "US",
		Region:      "CA",
		ExpiresAt:   time.Now().Add(-time.Minute),
	})
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return responseWithBody(`{"countryCode":"DE","region":"BE","status":"success"}`), nil
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	country, region, cacheMarker, err := cfg.GetCountryCode(context.Background(), "1.2.3.4")

	assert.NoError(t, err)
	assert.Equal(t, "DE", country)
	assert.Equal(t, "BE", region)
	assert.Equal(t, "-", cacheMarker)
	assert.Equal(t, 1, client.calls)

	cached, found := tempCache.Get("1.2.3.4")
	assert.True(t, found)
	assert.Equal(t, "DE", cached.CountryCode)
	assert.False(t, cached.ExpiresAt.IsZero())
	assert.True(t, cached.ExpiresAt.After(time.Now()))
}

func TestGetCountryCodeCacheMissError(t *testing.T) {
	// Initialize a new cache for this test
	tempCache := newTestCache(t, 16)
	IPCache = tempCache // Set the global cache for this test

	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &GetCountryCodeConfig{HTTPClient: client}

	_, _, cacheMarker, err := cfg.GetCountryCode(context.Background(), "1.2.3.4")

	assert.Error(t, err)
	assert.Equal(t, "-", cacheMarker)
	_, found := tempCache.Get("1.2.3.4")
	assert.False(t, found) // Should not cache on error
}

func TestGetCountryCodeCachesFailuresWithTTL(t *testing.T) {
	tempCache := newTestCache(t, 16)
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &GetCountryCodeConfig{
		HTTPClient: client,
		Cache:      tempCache,
		FailureTTL: time.Minute,
	}

	_, _, marker, err := cfg.GetCountryCode(context.Background(), "1.2.3.4")
	assert.Error(t, err)
	assert.Equal(t, "-", marker)
	assert.Equal(t, 1, client.calls)

	_, _, marker, err = cfg.GetCountryCode(context.Background(), "1.2.3.4")
	assert.Error(t, err)
	assert.Equal(t, "cached-failure", marker)
	assert.Equal(t, 1, client.calls)
}

func TestGetCountryCodeExpiredFailureCacheRetries(t *testing.T) {
	tempCache := newTestCache(t, 16)
	client := &mockHTTPClient{
		getFunc: func(_ context.Context, url string) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	}
	cfg := &GetCountryCodeConfig{
		HTTPClient: client,
		Cache:      tempCache,
		FailureTTL: 2 * time.Millisecond,
	}

	_, _, _, err := cfg.GetCountryCode(context.Background(), "1.2.3.4")
	assert.Error(t, err)
	assert.Equal(t, 1, client.calls)

	time.Sleep(5 * time.Millisecond)

	_, _, _, err = cfg.GetCountryCode(context.Background(), "1.2.3.4")
	assert.Error(t, err)
	assert.Equal(t, 2, client.calls)
}
