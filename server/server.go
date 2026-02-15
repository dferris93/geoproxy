package server

import (
	"context"
	"fmt"
	"geoproxy/common"
	"geoproxy/handler"
	"geoproxy/ipapi"
	"log"
	"net"
	"sync"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
)

type ClientHandlerFactory interface {
	NewClientHandler() handler.Handler
}
type HandlerFactory struct {
	AllowedCountries     map[string]bool
	AllowedRegions       map[string]bool
	DeniedCountries      map[string]bool
	DeniedRegions        map[string]bool
	AlwaysAllowed        []string
	AlwaysDenied         []string
	IPApiClient          ipapi.IPAPI
	CheckIps             common.CheckIP
	TransferFunc         func(handler.Connection, handler.Connection, *proxyproto.Header)
	BackendDialer        handler.BackendDialer
	BackendIP            string
	BackendPort          string
	SendProxyProtocol    bool
	ProxyProtocolVersion int
	MaxConnLifetime      time.Duration
	StartTime            time.Time
	EndTime              time.Time
	StartDate            time.Time
	EndDate              time.Time
	DaysOfWeek           map[time.Weekday]bool
	IdleTimeout          time.Duration
	ConnLimiter          handler.ConnLimiter
}

func (h *HandlerFactory) NewClientHandler() handler.Handler {
	return &handler.ClientHandler{
		AllowedCountries:     h.AllowedCountries,
		AllowedRegions:       h.AllowedRegions,
		DeniedCountries:      h.DeniedCountries,
		DeniedRegions:        h.DeniedRegions,
		AlwaysAllowed:        h.AlwaysAllowed,
		AlwaysDenied:         h.AlwaysDenied,
		IPApiClient:          h.IPApiClient,
		CheckIps:             h.CheckIps,
		TransferFunc:         h.TransferFunc,
		BackendDialer:        h.BackendDialer,
		BackendAddr:          h.BackendIP,
		BackendPort:          h.BackendPort,
		SendProxyProtocol:    h.SendProxyProtocol,
		ProxyProtocolVersion: h.ProxyProtocolVersion,
		MaxConnLifetime:      h.MaxConnLifetime,
		StartTime:            h.StartTime,
		EndTime:              h.EndTime,
		StartDate:            h.StartDate,
		EndDate:              h.EndDate,
		DaysOfWeek:           h.DaysOfWeek,
		IdleTimeout:          h.IdleTimeout,
		ConnLimiter:          h.ConnLimiter,
	}
}

type ServerConfig struct {
	ListenIP          string
	ListenPort        string
	BackendIP         string
	BackendPort       string
	NetListener       NetListener
	HandlerFactory    ClientHandlerFactory
	serverErrorMu     sync.Mutex
	serverError       error
	RecvProxyProtocol bool
	TrustedProxies    []string
	MaxConns          int
	ProxyProtoTimeout time.Duration
}

func (s *ServerConfig) StartServer(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	listenAddr := fmt.Sprintf("%s:%s", s.ListenIP, s.ListenPort)

	if s.RecvProxyProtocol && len(s.TrustedProxies) == 0 {
		s.setServerError(fmt.Errorf("recvProxyProtocol enabled but trustedProxies is empty"))
		log.Printf("failed to start server on %s: %v", listenAddr, s.ServerError())
		return
	}

	l, err := s.NetListener.Listen("tcp", listenAddr)
	if err != nil {
		s.setServerError(err)
		log.Printf("failed to start tcp server on %s: %v", listenAddr, err)
		return
	}

	if s.RecvProxyProtocol {
		timeout := s.ProxyProtoTimeout
		if timeout <= 0 {
			timeout = 1 * time.Second
		}
		proxyListener := &proxyproto.Listener{Listener: l, ReadHeaderTimeout: timeout}
		allowed, err := compileTrustedProxies(s.TrustedProxies)
		if err != nil {
			s.setServerError(err)
			log.Printf("failed to configure proxy protocol policy on %s: %v", listenAddr, err)
			return
		}
		// Reject non-trusted upstreams during Accept (no PROXY header parsing).
		// Require a PROXY header from trusted upstreams to avoid misconfig that
		// accidentally geofilters on the proxy's IP.
		proxyListener.Policy = func(upstream net.Addr) (proxyproto.Policy, error) {
			ip := ipFromAddr(upstream)
			if ip == nil {
				return proxyproto.REJECT, proxyproto.ErrInvalidUpstream
			}
			if !allowed(ip) {
				return proxyproto.REJECT, proxyproto.ErrInvalidUpstream
			}
			return proxyproto.REQUIRE, nil
		}
		l = proxyListener
	}

	listener := &listener{Listener: l}
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	var sem chan struct{}
	if s.MaxConns > 0 {
		sem = make(chan struct{}, s.MaxConns)
	}

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			s.setServerError(err)
			log.Printf("failed to accept connection: %v", err)
			err = checkCanceled(ctx)
			if err != nil {
				log.Printf("shutting down server on %s", listenAddr)
				listener.Close()
				return
			}
			continue
		}

		if sem != nil {
			select {
			case sem <- struct{}{}:
			default:
				// Avoid touching proxyproto.Conn.RemoteAddr() here (can block on PROXY header read).
				log.Printf("too many active connections on %s; rejecting connection", listenAddr)
				_ = clientConn.Close()
				continue
			}
		}

		// Do not log clientConn.RemoteAddr() here when recvProxyProtocol is enabled:
		// proxyproto.Conn.RemoteAddr() will attempt to read the PROXY header and can
		// block this accept loop (slowloris/DoS).

		handler := s.HandlerFactory.NewClientHandler()
		go func() {
			if sem != nil {
				defer func() { <-sem }()
			}
			handler.HandleClient(ctx, clientConn)
		}()
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

func (s *ServerConfig) setServerError(err error) {
	s.serverErrorMu.Lock()
	defer s.serverErrorMu.Unlock()
	s.serverError = err
}

func (s *ServerConfig) ServerError() error {
	s.serverErrorMu.Lock()
	defer s.serverErrorMu.Unlock()
	return s.serverError
}

func compileTrustedProxies(entries []string) (func(net.IP) bool, error) {
	var ips []net.IP
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(entry); err == nil {
			return nil, fmt.Errorf("CIDRs are not allowed in trustedProxies (got %q)", entry)
		}
		ip := net.ParseIP(entry)
		if ip == nil {
			return nil, fmt.Errorf("invalid trusted proxy entry %q", entry)
		}
		ips = append(ips, ip)
	}
	return func(ip net.IP) bool {
		for _, a := range ips {
			if a.Equal(ip) {
				return true
			}
		}
		return false
	}, nil
}

func ipFromAddr(a net.Addr) net.IP {
	if a == nil {
		return nil
	}
	if ta, ok := a.(*net.TCPAddr); ok && ta.IP != nil {
		return ta.IP
	}
	host, _, err := net.SplitHostPort(a.String())
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}
