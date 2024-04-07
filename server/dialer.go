package server

import (
	"net"
)

type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

type RealDialer struct{}

func (d *RealDialer) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}