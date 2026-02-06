package handler

import (
	"context"
	"net"
)

// BackendDialer abstracts backend dialing so handlers can dial only after an allow decision.
type BackendDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

