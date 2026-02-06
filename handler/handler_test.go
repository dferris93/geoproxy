package handler

import (
	"context"
	"fmt"
	"geoproxy/common"
	"geoproxy/mocks"
	"net"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
)

func TransferFuncMock(ClientConn Connection, BackendConn Connection, proxyHeader *proxyproto.Header) {
}

type staticDialer struct {
	conn net.Conn
	err  error
}

func (d *staticDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.conn, d.err
}

func TestHandler(t *testing.T) {
	t.Run("TestAlwaysAllowedv4True", func(t *testing.T) {
		fmt.Println("TestAlwaysAllowedv4True")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{"127.0.0.1"},
			AlwaysDenied:     []string{},
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendDialer:    &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:      "127.0.0.1",
			BackendPort:      "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestAlwaysAllowedv6True", func(t *testing.T) {
		fmt.Println("TestAlwaysAllowedv6True")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{"fe80::3c9e:f7ff:febc:caa"},
			AlwaysDenied:     []string{},
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},

			CheckIps:      &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 6}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 6}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestAlwaysAllowedFalse", func(t *testing.T) {
		fmt.Println("TestAlwaysAllowedFalse")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{"127.0.0.1"},
			AlwaysDenied:     []string{},
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},

			CheckIps:      &MockCheckIP{CheckSubnetsReturn: false, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAlwaysDeniedv4True", func(t *testing.T) {
		fmt.Println("TestAlwaysDeniedv4True")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{"127.0.0.1"},
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},

			CheckIps:      &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAlwaysDeniedv6True", func(t *testing.T) {
		fmt.Println("TestAlwaysDeniedv6True")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{"fe80::3c9e:f7ff:febc:caa"},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},

			CheckIps:      &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 6}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 6}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAlwaysDeniedFalse", func(t *testing.T) {
		fmt.Println("TestAlwaysDeniedFalse")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{"127.0.0.1"},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},

			CheckIps:      &MockCheckIP{CheckSubnetsReturn: false, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestDeniedCountries", func(t *testing.T) {
		fmt.Println("TestDeniedCountries")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{"CN": true},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAllowedCountries", func(t *testing.T) {
		fmt.Println("TestAllowedCountries")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestAllowedRegions", func(t *testing.T) {
		fmt.Println("TestAllowedRegions")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{"Beijing": true},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestDeniedRegions", func(t *testing.T) {
		fmt.Println("TestDeniedRegions")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{"Beijing": true},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("Test Start and End Time is good before midnight", func(t *testing.T) {
		fmt.Println("TestGoodBeforeMidnight")
		startTime, _ := time.Parse("15:04", "23:58")
		endTime, _ := time.Parse("15:04", "01:00")
		now, _ := time.Parse("15:04", "23:59")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			StartTime:     startTime,
			EndTime:       endTime,
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("Test Start and End Time is good past midnight", func(t *testing.T) {
		fmt.Println("TestgoodPastMidnight")
		startTime, _ := time.Parse("15:04", "23:59")
		endTime, _ := time.Parse("15:04", "01:00")
		now, _ := time.Parse("15:04", "00:01")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			StartTime:     startTime,
			EndTime:       endTime,
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("Test Start and End Time is bad before start", func(t *testing.T) {
		fmt.Println("TestBadBeforeStart")
		now, _ := time.Parse("15:04", "23:58")
		startTime, _ := time.Parse("15:04", "23:59")
		endTime, _ := time.Parse("15:04", "01:00")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			StartTime:     startTime,
			EndTime:       endTime,
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("Test Start and End Time is bad after end", func(t *testing.T) {
		fmt.Println("TestBadAfterEnd")
		now, _ := time.Parse("15:04", "01:01")
		startTime, _ := time.Parse("15:04", "23:59")
		endTime, _ := time.Parse("15:04", "01:00")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			StartTime:     startTime,
			EndTime:       endTime,
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("Test Start and End Date is good", func(t *testing.T) {
		fmt.Println("TestGoodDateRange")
		startDate, _ := time.Parse("2006-01-02", "2024-08-01")
		endDate, _ := time.Parse("2006-01-02", "2024-08-31")
		now, _ := time.Parse("2006-01-02", "2024-08-15")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			StartDate:     startDate,
			EndDate:       endDate,
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("Test Start and End Date is bad", func(t *testing.T) {
		fmt.Println("TestBadDateRange")
		startDate, _ := time.Parse("2006-01-02", "2024-08-01")
		endDate, _ := time.Parse("2006-01-02", "2024-08-31")
		now, _ := time.Parse("2006-01-02", "2024-09-01")
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			StartDate:     startDate,
			EndDate:       endDate,
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("Test Days of Week is good", func(t *testing.T) {
		fmt.Println("TestGoodDaysOfWeek")
		now := time.Date(2024, time.September, 2, 10, 0, 0, 0, time.UTC)
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			DaysOfWeek:    map[time.Weekday]bool{time.Monday: true},
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("Test Days of Week is bad", func(t *testing.T) {
		fmt.Println("TestBadDaysOfWeek")
		now := time.Date(2024, time.September, 2, 10, 0, 0, 0, time.UTC)
		h := ClientHandler{
			AllowedCountries: map[string]bool{"CN": true},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},

			IPApiClient: &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},

			CheckIps:      &common.CheckIPs{},
			TransferFunc:  TransferFuncMock,
			BackendDialer: &staticDialer{conn: &mocks.MockNetConn{IPVersion: 4}},
			BackendAddr:   "127.0.0.1",
			BackendPort:   "8080",
			DaysOfWeek:    map[time.Weekday]bool{time.Tuesday: true},
			Now:           now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(context.Background(), &ClientConn)
		assert.Equal(t, false, h.accepted)
	})
}
