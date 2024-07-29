package server

import (
	"net"
)
type NetListener interface {
	Listen(network, address string) (net.Listener, error)
}

type RealNetListener struct{}

func (rnl *RealNetListener) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
}

type listener struct {
	Listener net.Listener
}

func (rl *listener) Accept() (net.Conn, error) {
	return rl.Listener.Accept()
}

func (rl *listener) Close() error {
	return rl.Listener.Close()
}

func (rl *listener) Addr() net.Addr {
	return rl.Listener.Addr()
}