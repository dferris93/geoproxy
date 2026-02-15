package handler

import (
	"context"
	"geoproxy/common"
	"geoproxy/ipapi"
	"io"
	"log"
	"net"
	"strings"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
)

type Handler interface {
	HandleClient(context.Context, Connection)
}

type ClientHandler struct {
	AllowedCountries     map[string]bool
	AllowedRegions       map[string]bool
	DeniedCountries      map[string]bool
	DeniedRegions        map[string]bool
	AlwaysAllowed        []string
	AlwaysDenied         []string
	IPApiClient          ipapi.IPAPI
	CheckIps             common.CheckIP
	TransferFunc         func(Connection, Connection, *proxyproto.Header)
	BackendDialer        BackendDialer
	BackendAddr          string
	BackendPort          string
	countryCode          string
	region               string
	cached               string
	clientConn           Connection
	accepted             bool
	clientAddr           string
	clientIP             string
	SendProxyProtocol    bool
	ProxyProtocolVersion int
	MaxConnLifetime      time.Duration
	StartTime            time.Time
	EndTime              time.Time
	StartDate            time.Time
	EndDate              time.Time
	DaysOfWeek           map[time.Weekday]bool
	Now                  time.Time
	DeniedReason         string
	IdleTimeout          time.Duration
	ConnLimiter          ConnLimiter
}

func (h *ClientHandler) HandleClient(ctx context.Context, ClientConn Connection) {
	// Force early PROXY header parsing (if any) so we can reject invalid headers
	// before doing ip-api work or dialing the backend.
	if _, err := ClientConn.Read(make([]byte, 0)); err != nil {
		log.Printf("Failed to read/validate PROXY header: %v", err)
		_ = ClientConn.Close()
		return
	}

	clientAddr := ClientConn.RemoteAddr().String()

	ip, _, err := net.SplitHostPort(clientAddr)
	if err != nil {
		log.Printf("Failed to split host port: %v", err)
		_ = ClientConn.Close()
		return
	}
	if i := strings.LastIndexByte(ip, '%'); i >= 0 {
		ip = ip[:i]
	}

	h.clientAddr = clientAddr
	h.clientIP = ip
	h.countryCode = "--"
	h.region = "--"
	h.cached = "--"
	h.clientConn = ClientConn

	if h.ConnLimiter != nil {
		if !h.ConnLimiter.Acquire(ip) {
			h.accepted = false
			h.DeniedReason = "too many concurrent connections from source IP"
			h.processConnection(ctx)
			return
		}
		defer h.ConnLimiter.Release(ip)
	}

	if len(h.AlwaysDenied) > 0 {
		if h.CheckIps.CheckSubnets(h.AlwaysDenied, ip) {
			h.accepted = false
			h.DeniedReason = "Always denied"
			h.processConnection(ctx)
			return
		}
	}

	if len(h.AlwaysAllowed) > 0 {
		if h.CheckIps.CheckSubnets(h.AlwaysAllowed, ip) {
			h.accepted = true
			h.processConnection(ctx)
			return
		}
	}

	now := h.Now
	if now.IsZero() {
		now = time.Now()
	}
	if (!h.StartDate.IsZero() && !h.EndDate.IsZero()) || (!h.StartTime.IsZero() && !h.EndTime.IsZero()) || len(h.DaysOfWeek) > 0 {
		if !h.StartDate.IsZero() && !h.EndDate.IsZero() {
			ok, err := common.CheckDateRange(h.StartDate, h.EndDate, now)
			if err != nil {
				log.Printf("Failed to check date range: %v", err)
				_ = ClientConn.Close()
				return
			}
			if !ok {
				h.accepted = false
				h.DeniedReason = "connection not allowed on this date"
				h.processConnection(ctx)
				return
			}
		}
		if len(h.DaysOfWeek) > 0 {
			if !h.DaysOfWeek[now.Weekday()] {
				h.accepted = false
				h.DeniedReason = "connection not allowed on this day"
				h.processConnection(ctx)
				return
			}
		}
		if !h.StartTime.IsZero() && !h.EndTime.IsZero() {
			ok, err := common.CheckTime(h.StartTime, h.EndTime, now)
			if err != nil {
				log.Printf("Failed to check time: %v", err)
				_ = ClientConn.Close()
				return
			}
			if !ok {
				h.accepted = false
				h.DeniedReason = "connection not allowed at this time"
				h.processConnection(ctx)
				return
			}
		}
	}

	h.countryCode, h.region, h.cached, err = h.IPApiClient.GetCountryCode(ctx, ip)
	if err != nil {
		log.Printf("ipapi connection error: %v", err)
		h.accepted = false
		h.DeniedReason = "ipapi error"
		h.processConnection(ctx)
		return
	}
	normalizedCountry := strings.ToUpper(strings.TrimSpace(h.countryCode))
	normalizedRegion := strings.ToUpper(strings.TrimSpace(h.region))

	countryAccepted := false
	regionAccepted := true

	if _, ok := h.DeniedCountries[normalizedCountry]; ok {
		countryAccepted = false
	} else {
		if _, ok := h.AllowedCountries[normalizedCountry]; ok {
			countryAccepted = true
			if len(h.DeniedRegions) > 0 {
				if _, ok := h.DeniedRegions[normalizedRegion]; ok {
					regionAccepted = false
				}
			} else {
				if len(h.AllowedRegions) > 0 {
					if _, ok := h.AllowedRegions[normalizedRegion]; !ok {
						regionAccepted = false
					}
				}
			}
		}
	}

	h.accepted = countryAccepted && regionAccepted
	if !h.accepted {
		h.DeniedReason = "country or region denied"
	}
	h.processConnection(ctx)
}

func (h *ClientHandler) processConnection(ctx context.Context) {
	if h.accepted {
		if h.BackendDialer == nil {
			log.Printf("no backend dialer configured; dropping connection from %s", h.clientAddr)
			_ = h.clientConn.Close()
			return
		}
		backendTuple := net.JoinHostPort(h.BackendAddr, h.BackendPort)
		backendConn, err := h.BackendDialer.DialContext(ctx, "tcp", backendTuple)
		if err != nil {
			log.Printf("failed to connect to backend %s: %v", backendTuple, err)
			_ = h.clientConn.Close()
			return
		}

		log.Printf("accepted connection from %s country: %s region: %s to %s:%s %s",
			h.clientAddr,
			h.countryCode,
			h.region,
			h.BackendAddr,
			h.BackendPort,
			h.cached)
		clientConn := withConnLimits(h.clientConn, h.IdleTimeout, h.MaxConnLifetime)
		backendWrapped := withConnLimits(Connection(backendConn), h.IdleTimeout, h.MaxConnLifetime)

		var hdr *proxyproto.Header
		if h.SendProxyProtocol {
			hdr = proxyproto.HeaderProxyFromAddrs(byte(h.ProxyProtocolVersion), clientConn.RemoteAddr(), backendConn.RemoteAddr())
		}
		h.TransferFunc(clientConn, backendWrapped, hdr)

		log.Printf("closed connection from %s country: %s region: %s to %s:%s %s",
			h.clientAddr,
			h.countryCode,
			h.region,
			h.BackendAddr,
			h.BackendPort,
			h.cached)
	} else {
		log.Printf("rejected connection from %s country: %s region: %s to %s:%s %s reason: %s",
			h.clientAddr,
			h.countryCode,
			h.region,
			h.BackendAddr,
			h.BackendPort,
			h.cached,
			h.DeniedReason)
		_ = h.clientConn.Close()
	}
}

func TransferData(ClientConn Connection, BackendConn Connection, h *proxyproto.Header) {
	defer func() { _ = ClientConn.Close() }()
	defer func() { _ = BackendConn.Close() }()

	// Ensure PROXY header is fully written before forwarding any client bytes.
	if h != nil {
		if _, err := h.WriteTo(BackendConn); err != nil {
			log.Printf("Error writing proxy header: %v", err)
			return
		}
	}

	type closeWriter interface{ CloseWrite() error }
	type tcpConnGetter interface {
		TCPConn() (*net.TCPConn, bool)
	}
	type rawConnGetter interface {
		Raw() net.Conn
	}

	closeWriteOrClose := func(c Connection) {
		if cw, ok := c.(closeWriter); ok {
			_ = cw.CloseWrite()
			return
		}
		// Some wrappers (notably proxyproto.Conn) don't expose CloseWrite, but can
		// provide the underlying TCPConn.
		if tg, ok := c.(tcpConnGetter); ok {
			if tcp, ok := tg.TCPConn(); ok {
				_ = tcp.CloseWrite()
				return
			}
		}
		if rg, ok := c.(rawConnGetter); ok {
			if cw, ok := rg.Raw().(closeWriter); ok {
				_ = cw.CloseWrite()
				return
			}
		}
		_ = c.Close()
	}

	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(BackendConn, ClientConn)
		// Signal EOF to backend (or fully close if half-close isn't available).
		closeWriteOrClose(BackendConn)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(ClientConn, BackendConn)
		// Signal EOF to client (or fully close if half-close isn't available).
		closeWriteOrClose(ClientConn)
		errCh <- err
	}()

	err1 := <-errCh
	err2 := <-errCh
	if err1 != nil && err1 != io.EOF {
		log.Printf("Error copying data (client->backend): %v", err1)
	}
	if err2 != nil && err2 != io.EOF {
		log.Printf("Error copying data (backend->client): %v", err2)
	}
}
