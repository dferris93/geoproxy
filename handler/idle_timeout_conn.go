package handler

import (
	"net"
	"time"
)

type limitConn struct {
	Connection
	idle         time.Duration
	hardDeadline time.Time
}

func withConnLimits(c Connection, idle time.Duration, maxLifetime time.Duration) Connection {
	if c == nil {
		return nil
	}
	var hard time.Time
	if maxLifetime > 0 {
		hard = time.Now().Add(maxLifetime)
	}
	if idle <= 0 && hard.IsZero() {
		return c
	}
	return &limitConn{Connection: c, idle: idle, hardDeadline: hard}
}

func (c *limitConn) Read(b []byte) (int, error) {
	_ = c.Connection.SetReadDeadline(c.nextDeadline(time.Now(), true))
	return c.Connection.Read(b)
}

func (c *limitConn) Write(b []byte) (int, error) {
	_ = c.Connection.SetWriteDeadline(c.nextDeadline(time.Now(), false))
	return c.Connection.Write(b)
}

func (c *limitConn) nextDeadline(now time.Time, _ bool) time.Time {
	var d time.Time
	if c.idle > 0 {
		d = now.Add(c.idle)
	}
	if c.hardDeadline.IsZero() {
		return d
	}
	if d.IsZero() || c.hardDeadline.Before(d) {
		return c.hardDeadline
	}
	return d
}

// Optional TCP half-close support (used by TransferData to avoid truncation).
func (c *limitConn) CloseWrite() error {
	if cw, ok := c.Connection.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	if tg, ok := c.Connection.(interface{ TCPConn() (*net.TCPConn, bool) }); ok {
		if tcp, ok := tg.TCPConn(); ok {
			return tcp.CloseWrite()
		}
	}
	if rg, ok := c.Connection.(interface{ Raw() net.Conn }); ok {
		if cw, ok := rg.Raw().(interface{ CloseWrite() error }); ok {
			return cw.CloseWrite()
		}
	}
	return c.Connection.Close()
}

func (c *limitConn) CloseRead() error {
	if cr, ok := c.Connection.(interface{ CloseRead() error }); ok {
		return cr.CloseRead()
	}
	if tg, ok := c.Connection.(interface{ TCPConn() (*net.TCPConn, bool) }); ok {
		if tcp, ok := tg.TCPConn(); ok {
			return tcp.CloseRead()
		}
	}
	if rg, ok := c.Connection.(interface{ Raw() net.Conn }); ok {
		if cr, ok := rg.Raw().(interface{ CloseRead() error }); ok {
			return cr.CloseRead()
		}
	}
	return c.Connection.Close()
}
