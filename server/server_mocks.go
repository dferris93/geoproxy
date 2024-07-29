package server

import (
	"fmt"
	"geoproxy/handler"
	"geoproxy/mocks"
	"net"

	"github.com/pires/go-proxyproto"
)

type MockNetListener struct {
	SendError           bool
	ListenerAcceptError bool
	ListenerCloseError  bool
	ListenerAddrError   bool
}

func (m *MockNetListener) Listen(network, address string) (net.Listener, error) {
	if m.SendError {
		return nil, fmt.Errorf("NetListener error")
	} else {
		return &MockListener{
			AcceptError: m.ListenerAcceptError,
			CloseError:  m.ListenerCloseError,
			AddrError:   m.ListenerAddrError,
		}, nil
	}
}

type MockListener struct {
	AcceptError bool
	CloseError  bool
	AddrError   bool
}

func (m *MockListener) Accept() (net.Conn, error) {
	if m.AcceptError {
		return nil, fmt.Errorf("listener error")
	} else {
		return &mocks.MockNetConn{}, nil
	}
}

func (m *MockListener) Close() error {
	if m.CloseError {
		return fmt.Errorf("Listener error")
	} else {
		return nil
	}
}

func (m *MockListener) Addr() net.Addr {
	return &mocks.MockNetAddr{}
}

type MockNetDialer struct {
	DialError bool
}

func (m *MockNetDialer) Dial(network, address string) (net.Conn, error) {
	if m.DialError {
		return nil, fmt.Errorf("Dialer error")
	} else {
		return &mocks.MockNetConn{}, nil
	}
}

type MockHandlerFactory struct {}

func (m *MockHandlerFactory) NewClientHandler() handler.Handler {
	return &MockClientHandler{}
}

type MockClientHandler struct {}

func (m *MockClientHandler) HandleClient(ClientConn handler.Connection, BackendConn handler.Connection, proxyHeader *proxyproto.Header) {}

