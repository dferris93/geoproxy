package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"geoproxy/handler"

	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
)

type captureHandler struct {
	headerCh chan *proxyproto.Header
}

func (c *captureHandler) HandleClient(client handler.Connection, backend handler.Connection, hdr *proxyproto.Header) {
	c.headerCh <- hdr
}

type captureHandlerFactory struct {
	handler *captureHandler
}

func (c *captureHandlerFactory) NewClientHandler() handler.Handler {
	return c.handler
}

func TestHandlerFactoryNewClientHandler(t *testing.T) {
	startTime := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.January, 1, 17, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)
	days := map[time.Weekday]bool{time.Monday: true}

	factory := &HandlerFactory{
		AllowedCountries: map[string]bool{"US": true},
		AllowedRegions:   map[string]bool{"CA": true},
		DeniedCountries:  map[string]bool{"CN": true},
		DeniedRegions:    map[string]bool{"BJ": true},
		AlwaysAllowed:    []string{"127.0.0.1"},
		AlwaysDenied:     []string{"10.0.0.1"},
		ContinueOnError:  true,
		CheckIps:         nil,
		TransferFunc:     nil,
		BackendIP:        "127.0.0.1",
		BackendPort:      "8080",
		StartTime:        startTime,
		EndTime:          endTime,
		StartDate:        startDate,
		EndDate:          endDate,
		DaysOfWeek:       days,
	}

	h := factory.NewClientHandler()
	clientHandler, ok := h.(*handler.ClientHandler)
	if assert.True(t, ok) {
		assert.Equal(t, factory.AllowedCountries, clientHandler.AllowedCountries)
		assert.Equal(t, factory.AllowedRegions, clientHandler.AllowedRegions)
		assert.Equal(t, factory.DeniedCountries, clientHandler.DeniedCountries)
		assert.Equal(t, factory.DeniedRegions, clientHandler.DeniedRegions)
		assert.Equal(t, factory.AlwaysAllowed, clientHandler.AlwaysAllowed)
		assert.Equal(t, factory.AlwaysDenied, clientHandler.AlwaysDenied)
		assert.Equal(t, factory.ContinueOnError, clientHandler.ContinueOnError)

		assert.Equal(t, factory.BackendIP, clientHandler.BackendAddr)
		assert.Equal(t, factory.BackendPort, clientHandler.BackendPort)
		assert.Equal(t, factory.StartTime, clientHandler.StartTime)
		assert.Equal(t, factory.EndTime, clientHandler.EndTime)
		assert.Equal(t, factory.StartDate, clientHandler.StartDate)
		assert.Equal(t, factory.EndDate, clientHandler.EndDate)
		assert.Equal(t, factory.DaysOfWeek, clientHandler.DaysOfWeek)
	}
}

func TestStartServerSendProxyProtocol(t *testing.T) {
	headerCh := make(chan *proxyproto.Header, 1)
	handler := &captureHandler{headerCh: headerCh}
	factory := &captureHandlerFactory{handler: handler}

	s := &ServerConfig{
		ListenIP:             "127.0.0.1",
		ListenPort:           "8080",
		BackendIP:            "127.0.0.1",
		BackendPort:          "9090",
		NetListener:          &MockNetListener{},
		Dialer:               &MockNetDialer{},
		HandlerFactory:       factory,
		SendProxyProtocol:    true,
		ProxyProtocolVersion: 2,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.StartServer(&wg, ctx)
	wg.Wait()

	select {
	case header := <-headerCh:
		assert.NotNil(t, header)
		assert.Equal(t, byte(2), header.Version)
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected proxy protocol header to be sent")
	}
}
