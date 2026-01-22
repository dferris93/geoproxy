package handler

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestConnDelegates(t *testing.T) {
	left, right := net.Pipe()
	defer right.Close()

	conn := &Conn{Conn: left}

	go func() {
		_, _ = right.Write([]byte("hi"))
	}()

	buf := make([]byte, 2)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if n != 2 || string(buf) != "hi" {
		t.Fatalf("unexpected read: %q", string(buf[:n]))
	}

	readCh := make(chan []byte, 1)
	go func() {
		readBuf := make([]byte, 2)
		_, _ = io.ReadFull(right, readBuf)
		readCh <- readBuf
	}()
	_, err = conn.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	readBuf := <-readCh
	if string(readBuf) != "ok" {
		t.Fatalf("unexpected write: %q", string(readBuf))
	}

	if conn.LocalAddr().String() != left.LocalAddr().String() {
		t.Fatalf("LocalAddr mismatch")
	}
	if conn.RemoteAddr().String() != left.RemoteAddr().String() {
		t.Fatalf("RemoteAddr mismatch")
	}

	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetDeadline: %v", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("SetWriteDeadline: %v", err)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
