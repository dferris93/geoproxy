package handler

import (
	"fmt"
	"geoproxy/common"
	"geoproxy/mocks"
	"sync"
	"testing"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
)

func TransferFuncMock(ClientConn Connection, BackendConn Connection, proxyHeader *proxyproto.Header) {}

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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
		}
		ClientConn := mocks.MockNetConn{IPVersion: 6}
		BackendConn := mocks.MockNetConn{IPVersion: 6}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: false, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: true, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 6}
		BackendConn := mocks.MockNetConn{IPVersion: 6}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "US", ReturnRegion: "CA"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &MockCheckIP{CheckSubnetsReturn: false, CheckIPTypeReturn: 4, CheckIPTypeErr: false},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
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
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("Test Start and End Time is good before midnight", func(t *testing.T) {
		fmt.Println("TestGoodBeforeMidnight")
		startTime, _ := time.Parse("15:04", "23:58")
		endTime, _ := time.Parse("15:04", "01:00")
		now, _ := time.Parse("15:04", "23:59")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
			StartTime:       startTime,
			EndTime:         endTime,
			Now: 		  	 now,	 
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("Test Start and End Time is good past midnight", func(t *testing.T) {
		fmt.Println("TestgoodPastMidnight")
		startTime, _ := time.Parse("15:04", "23:59")
		endTime, _ := time.Parse("15:04", "01:00")
		now, _ := time.Parse("15:04", "00:01")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
			StartTime:       startTime,
			EndTime:         endTime,
			Now: 		  	 now,	 
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("Test Start and End Time is bad before start", func(t *testing.T) {
		fmt.Println("TestBadBeforeStart")
		now, _ := time.Parse("15:04", "23:58")
		startTime, _ := time.Parse("15:04", "23:59")
		endTime, _ := time.Parse("15:04", "01:00")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
			StartTime:       startTime,
			EndTime:         endTime,
			Now: 		  	 now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("Test Start and End Time is bad after end", func(t *testing.T) {
		fmt.Println("TestBadAfterEnd")
		now, _ := time.Parse("15:04", "01:01")
		startTime, _ := time.Parse("15:04", "23:59")
		endTime, _ := time.Parse("15:04", "01:00")
		h := ClientHandler{
			AllowedCountries: map[string]bool{},
			AllowedRegions:   map[string]bool{},
			DeniedCountries:  map[string]bool{},
			DeniedRegions:    map[string]bool{},
			AlwaysAllowed:    []string{},
			AlwaysDenied:     []string{},
			ContinueOnError:  false,
			IptablesBlock:    false,
			IPApiClient:      &GetCountryCodeMock{ReturnCountry: "CN", ReturnRegion: "Beijing"},
			Mutex:            &sync.Mutex{},
			BlockIPs:         make(chan string),
			CheckIps:         &common.CheckIPs{},
			TransferFunc:     TransferFuncMock,
			BackendAddr: 	 "127.0.0.1",
			BackendPort: 	 "8080",
			ProxyHeader: 	 nil,
			StartTime:       startTime,
			EndTime:         endTime,
			Now: 		  	 now,
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn, nil)
		assert.Equal(t, false, h.accepted)
	})
}