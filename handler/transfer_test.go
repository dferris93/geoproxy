package handler

import (
	"bytes"
	"net"
	"testing"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
)

func readUntilContains(conn net.Conn, needle string, closeOnMatch bool) []byte {
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	var buf []byte
	tmp := make([]byte, 64)
	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if bytes.Contains(buf, []byte(needle)) {
				if closeOnMatch {
					_ = conn.Close()
				}
				return buf
			}
		}
		if err != nil {
			return buf
		}
	}
}

func TestTransferDataWithProxyHeader(t *testing.T) {
	clientConn, clientPeer := net.Pipe()
	backendConn, backendPeer := net.Pipe()

	header := proxyproto.HeaderProxyFromAddrs(1,
		&net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 1234},
		&net.TCPAddr{IP: net.IPv4(5, 6, 7, 8), Port: 5678},
	)

	done := make(chan struct{})
	go func() {
		TransferData(clientConn, backendConn, header)
		close(done)
	}()

	backendReadCh := make(chan []byte, 1)
	go func() {
		backendReadCh <- readUntilContains(backendPeer, "ping", true)
	}()

	clientReadCh := make(chan []byte, 1)
	go func() {
		clientReadCh <- readUntilContains(clientPeer, "pong", false)
	}()

	if _, err := backendPeer.Write([]byte("pong")); err != nil {
		t.Fatalf("backend write: %v", err)
	}
	if _, err := clientPeer.Write([]byte("ping")); err != nil {
		t.Fatalf("client write: %v", err)
	}

	backendData := <-backendReadCh
	clientData := <-clientReadCh
	<-done

	if !bytes.Contains(backendData, []byte("PROXY")) {
		t.Fatalf("expected proxy header in backend data, got %q", string(backendData))
	}
	if !bytes.Contains(backendData, []byte("ping")) {
		t.Fatalf("expected ping in backend data, got %q", string(backendData))
	}
	if !bytes.Contains(clientData, []byte("pong")) {
		t.Fatalf("expected pong in client data, got %q", string(clientData))
	}
}
