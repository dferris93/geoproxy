package server

import (
	"context"
	"net"
	"time"
)

type RealDialer struct{
	Timeout time.Duration
}

func (d *RealDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	timeout := d.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	nd := net.Dialer{Timeout: timeout}
	return nd.DialContext(ctx, network, address)
}
