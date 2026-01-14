package handler

import (
	"geoproxy/common"
	"geoproxy/ipapi"
	"io"
	"log"
	"net"
	"sync"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
)

type Handler interface {
	HandleClient(Connection, Connection, *proxyproto.Header)
}

type ClientHandler struct {
	AllowedCountries map[string]bool
	AllowedRegions   map[string]bool
	DeniedCountries  map[string]bool
	DeniedRegions    map[string]bool
	AlwaysAllowed    []string
	AlwaysDenied     []string
	ContinueOnError  bool
	IptablesBlock    bool
	IPApiClient      ipapi.IPAPI
	Mutex            *sync.Mutex
	BlockIPs         chan string
	CheckIps         common.CheckIP
	TransferFunc     func(Connection, Connection, *proxyproto.Header)
	BackendAddr      string
	BackendPort      string
	countryCode      string
	region           string
	cached           string
	clientConn       Connection
	backendConn      Connection
	accepted         bool
	clientAddr       string
	ProxyHeader      *proxyproto.Header
	StartTime        time.Time
	EndTime          time.Time
	StartDate        time.Time
	EndDate          time.Time
	DaysOfWeek       map[time.Weekday]bool
	Now              time.Time
	DeniedReason     string
}

func (h *ClientHandler) HandleClient(ClientConn Connection, BackendConn Connection, proxyHeader *proxyproto.Header) {

	h.ProxyHeader = proxyHeader

	clientAddr := ClientConn.RemoteAddr().String()

	var err error

	ip, _, err := net.SplitHostPort(clientAddr)
	if err != nil {
		log.Printf("Failed to split host port: %v", err)
		return
	}

	h.clientAddr = clientAddr
	h.countryCode = "--"
	h.region = "--"
	h.cached = "--"
	h.clientConn = ClientConn
	h.backendConn = BackendConn

	if len(h.AlwaysDenied) > 0 {
		if h.CheckIps.CheckSubnets(h.AlwaysDenied, ip) {
			h.accepted = false
			h.DeniedReason = "Always denied"
			h.processConnection()
			return
		}
	}

	if len(h.AlwaysAllowed) > 0 {
		if h.CheckIps.CheckSubnets(h.AlwaysAllowed, ip) {
			h.accepted = true
			h.processConnection()
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
				return
			}
			if !ok {
				h.accepted = false
				h.DeniedReason = "connection not allowed on this date"
				h.processConnection()
				return
			}
		}
		if len(h.DaysOfWeek) > 0 {
			if !h.DaysOfWeek[now.Weekday()] {
				h.accepted = false
				h.DeniedReason = "connection not allowed on this day"
				h.processConnection()
				return
			}
		}
		if !h.StartTime.IsZero() && !h.EndTime.IsZero() {
			ok, err := common.CheckTime(h.StartTime, h.EndTime, now)
			if err != nil {
				log.Printf("Failed to check time: %v", err)
				return
			}
			if !ok {
				h.accepted = false
				h.DeniedReason = "connection not allowed at this time"
				h.processConnection()
				return
			}
		}
	}

	h.countryCode, h.region, h.cached, err = h.IPApiClient.GetCountryCode(ip, h.Mutex)
	if err != nil && !h.ContinueOnError {
		log.Printf("ipapi connection error: %v", err)
		if !h.ContinueOnError {
			ClientConn.Close()
			BackendConn.Close()
			return
		} else {
			log.Printf("continuing on despite error")
			h.accepted = true
			h.processConnection()
			return
		}
	}

	countryAccepted := false
	regionAccepted := true

	if _, ok := h.DeniedCountries[h.countryCode]; ok {
		countryAccepted = false
	} else {
		if _, ok := h.AllowedCountries[h.countryCode]; ok {
			countryAccepted = true
			if len(h.DeniedRegions) > 0 {
				if _, ok := h.DeniedRegions[h.region]; ok {
					regionAccepted = false
				}
			} else {
				if len(h.AllowedRegions) > 0 {
					if _, ok := h.AllowedRegions[h.region]; !ok {
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
	h.processConnection()
}

func (h *ClientHandler) processConnection() {
	if h.accepted {
		log.Printf("accepted connection from %s country: %s region: %s to %s:%s %s",
			h.clientAddr,
			h.countryCode,
			h.region,
			h.BackendAddr,
			h.BackendPort,
			h.cached)
		h.TransferFunc(h.clientConn, h.backendConn, h.ProxyHeader)
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
		if h.IptablesBlock {
			h.BlockIPs <- h.clientAddr
		}
	}
}

func TransferData(ClientConn Connection, BackendConn Connection, h *proxyproto.Header) {
	defer ClientConn.Close()
	defer BackendConn.Close()
	go func() {
		if h != nil {
			_, err := h.WriteTo(BackendConn)
			if err != nil {
				log.Printf("Error writing proxy header: %v", err)
			}
		}
		_, err := io.Copy(BackendConn, ClientConn)
		if err != nil {
			log.Printf("Error copying data from client to backend: %v", err)
		}
	}()

	_, err := io.Copy(ClientConn, BackendConn)
	if err != nil {
		log.Printf("Error copying data from backend to client: %v", err)
	}
}
