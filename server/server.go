package server

import (
	"context"
	"fmt"
	"geoproxy/common"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"log"
	"sync"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
)

type ClientHandlerFactory interface {
	NewClientHandler() handler.Handler
}
type HandlerFactory struct {
	AllowedCountries map[string]bool
	AllowedRegions   map[string]bool
	DeniedCountries  map[string]bool
	DeniedRegions    map[string]bool
	AlwaysAllowed    []string
	AlwaysDenied     []string
	ContinueOnError  bool
	IptablesBlock	bool
	IPApiClient    ipapi.IPAPI
	Mutex          *sync.Mutex
	BlockIPs	   chan string
	CheckIps     common.CheckIP
	TransferFunc func(handler.Connection, handler.Connection, *proxyproto.Header)
	BackendIP string
	BackendPort string
	StartTime time.Time
	EndTime time.Time
}

func (h *HandlerFactory) NewClientHandler() handler.Handler {
	return &handler.ClientHandler{
		AllowedCountries: h.AllowedCountries,
		AllowedRegions:   h.AllowedRegions,
		DeniedCountries:  h.DeniedCountries,
		DeniedRegions:    h.DeniedRegions,
		AlwaysAllowed:    h.AlwaysAllowed,
		AlwaysDenied:     h.AlwaysDenied,
		ContinueOnError: h.ContinueOnError,
		IptablesBlock:    h.IptablesBlock,
		IPApiClient: 	  h.IPApiClient,
		Mutex:            h.Mutex,
		BlockIPs:         h.BlockIPs,
		CheckIps:         h.CheckIps,
		TransferFunc:    h.TransferFunc,
		BackendAddr: 	h.BackendIP,
		BackendPort: 	h.BackendPort,
		ProxyHeader:    nil,
		StartTime:      h.StartTime,
		EndTime:        h.EndTime,
	}
}

type ServerConfig struct {
	ListenIP         string
	ListenPort       string
	BackendIP        string
	BackendPort      string
	NetListener      NetListener
	Dialer           Dialer
	HandlerFactory   ClientHandlerFactory
	serverError      error
	RecvProxyProtocol bool
	SendProxyProtocol bool
	ProxyProtocolVersion int
}

func (s *ServerConfig) StartServer(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	listenAddr := fmt.Sprintf("%s:%s", s.ListenIP, s.ListenPort)
	backendAddr := fmt.Sprintf("%s:%s", s.BackendIP, s.BackendPort)

	l, err := s.NetListener.Listen("tcp", listenAddr)
	if err != nil {
		s.serverError = err
		log.Printf("failed to start tcp server on %s: %v", listenAddr, err)
		return
	}

	if s.RecvProxyProtocol {
		l = &proxyproto.Listener{Listener: l, ReadHeaderTimeout: 5 * time.Second}
	}

	listener := &listener{Listener: l}

	handler := s.HandlerFactory.NewClientHandler()

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			s.serverError = err
			log.Printf("failed to accept connection: %v", err)
			err = checkCanceled(ctx)
			if err != nil {
				log.Printf("shutting down server on %s", listenAddr)
				listener.Close()
				return
			}
			continue
		}

		backendConn, err := s.Dialer.Dial("tcp", backendAddr)
		if err != nil {
			log.Printf("failed to connect to backend: %v", err)
			clientConn.Close()
			s.serverError = err
			err = checkCanceled(ctx)
			if err != nil {
				log.Printf("shutting down server on %s", listenAddr)
				listener.Close()
				return
			}
			continue
		}

		var proxyHeader *proxyproto.Header

		if s.SendProxyProtocol {
			proxyHeader = proxyproto.HeaderProxyFromAddrs(byte(s.ProxyProtocolVersion),
				clientConn.RemoteAddr(), backendConn.RemoteAddr())
		} else {
			proxyHeader = nil
		}

		log.Printf("connection from %s to %s:%s", clientConn.RemoteAddr(), s.ListenIP, s.ListenPort)
		go handler.HandleClient(clientConn, backendConn, proxyHeader)
		err = checkCanceled(ctx)
		if err != nil {
			log.Printf("shutting down server on %s", listenAddr)
			listener.Close()
			return
		}
	}
}

func checkCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}