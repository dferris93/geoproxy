package handler

import (
	"geoproxy/common"
	"geoproxy/ipapi"
	"io"
	"log"
	"net"
	"sync"

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
	ProxyHeader 	*proxyproto.Header
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

	h.countryCode, h.region, h.cached, err = h.IPApiClient.GetCountryCode(ip, h.Mutex)
	if err != nil && !h.ContinueOnError {
		log.Printf("ipapi connection error: %v", err)
		if !h.ContinueOnError {
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
		log.Printf("rejected connection from %s country: %s region: %s to %s:%s %s",
			h.clientAddr,
			h.countryCode,
			h.region,
			h.BackendAddr,
			h.BackendPort,
			h.cached)
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
