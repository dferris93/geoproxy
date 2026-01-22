package server

import (
	"context"
	"testing"
	"time"
)

func TestListenerWrapper(t *testing.T) {
	mock := &MockListener{}
	l := &listener{Listener: mock}

	if l.Addr() == nil {
		t.Fatal("expected non-nil addr")
	}
	if _, err := l.Accept(); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestCheckCanceled(t *testing.T) {
	if err := checkCanceled(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := checkCanceled(ctx); err == nil {
		t.Fatal("expected error after cancel")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Nanosecond)
	if err := checkCanceled(ctx); err == nil {
		t.Fatal("expected error after timeout")
	}
}
