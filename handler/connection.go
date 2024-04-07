package handler

import (
	"net"
	"time"
)

type Connection interface {
    RemoteAddr() net.Addr
	LocalAddr() net.Addr
    Close() error
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type Conn struct {
	Conn net.Conn
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Conn) Close() error {
	return c.Conn.Close()
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.Conn.Read(b)
}

func (c *Conn) Write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}
