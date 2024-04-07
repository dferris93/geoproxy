package handler

import (
	"geoproxy/common"
	"geoproxy/mocks"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TransferFuncMock(ClientConn Connection, BackendConn Connection) {}

func TestHandler(t *testing.T) {
	t.Run("TestAlwaysAllowedv4True", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestAlwaysAllowedv6True", func(t *testing.T) {
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
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestAlwaysAllowedFalse", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAlwaysDeniedv4True", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAlwaysDeniedv6True", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 6}
		BackendConn := mocks.MockNetConn{IPVersion: 6}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAlwaysDeniedFalse", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestDeniedCountries", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, false, h.accepted)
	})
	t.Run("TestAllowedCountries", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestAllowedRegions", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, true, h.accepted)
	})
	t.Run("TestDeniedRegions", func(t *testing.T) {
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
		}
		ClientConn := mocks.MockNetConn{IPVersion: 4}
		BackendConn := mocks.MockNetConn{IPVersion: 4}
		h.HandleClient(&ClientConn, &BackendConn)
		assert.Equal(t, false, h.accepted)
	})
}