package ipapi

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRealHTTPClientGetWithoutAPIKey(t *testing.T) {
	var gotPath string
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})
	defer func() { http.DefaultTransport = oldTransport }()

	client := &RealHTTPClient{Endpoint: "http://example.test/json/"}
	resp, err := client.Get(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()

	if gotPath != "/json/1.2.3.4" {
		t.Fatalf("expected path /json/1.2.3.4, got %s", gotPath)
	}
}

func TestRealHTTPClientGetWithAPIKey(t *testing.T) {
	var gotPath string
	var gotQuery string
	var gotAPIKey string
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAPIKey = r.Header.Get("X-API-Key")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
			Header:     make(http.Header),
		}, nil
	})
	defer func() { http.DefaultTransport = oldTransport }()

	client := &RealHTTPClient{Endpoint: "http://example.test/json", APIKey: "testkey"}
	resp, err := client.Get(context.Background(), "1.2.3.4")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()

	if gotPath != "/json/1.2.3.4" {
		t.Fatalf("expected path /json/1.2.3.4, got %s", gotPath)
	}
	q, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	if q.Get("key") != "" {
		t.Fatalf("expected key query param to be empty, got: %s", q.Get("key"))
	}
	if q.Get("fields") != "countryCode,region,status" {
		t.Fatalf("unexpected fields: %s", q.Get("fields"))
	}
	if gotAPIKey != "testkey" {
		t.Fatalf("unexpected X-API-Key header: %s", gotAPIKey)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
