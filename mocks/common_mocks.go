package mocks

import (
	"net"
	"time"
)

type MockNetConn struct {
	ReadError             bool
	WriteError            bool
	CloseError            bool
	SetDeadlineError      bool
	SetReadDeadlineError  bool
	SetWriteDeadlineError bool
	IPVersion             int
}

func (m *MockNetConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockNetConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockNetConn) Close() error {
	return nil
}

func (m *MockNetConn) LocalAddr() net.Addr {
	if m.IPVersion == 4 {
		return &MockNetAddr{IPVersion: 4}
	} else {
		return &MockNetAddr{IPVersion: 6}
	}

}

func (m *MockNetConn) RemoteAddr() net.Addr {
	if m.IPVersion == 4 {
		return &MockNetAddr{IPVersion: 4}
	} else {
		return &MockNetAddr{IPVersion: 6}
	}
}

func (m *MockNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type MockNetAddr struct{
	IPVersion int
}

func (m *MockNetAddr) Network() string {
	return "tcp"
}

func (m *MockNetAddr) String() string {
	if m.IPVersion == 4 {
		return "127.0.0.1:8080"
	} else {
		return "[fe80::3c9e:f7ff:febc:caa]:8080"
	}
}