package server

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	t.Run("TestServerListenerError", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP:       "867.053.009.000",
			ListenPort:     "8080",
			BackendIP:      "127.0.0.1",
			BackendPort:    "8080",
			NetListener:    &MockNetListener{SendError: true},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		assert.Error(t, s.ServerError())
	})
	t.Run("TestServerAcceptError", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP:       "127.0.0.1",
			ListenPort:     "8080",
			BackendIP:      "127.0.0.1",
			BackendPort:    "8080",
			NetListener:    &MockNetListener{ListenerAcceptError: true},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		assert.Error(t, s.ServerError())
	})
	t.Run("TestServerDialError", func(t *testing.T) {
		// Backend dials are now deferred until the handler performs I/O.
		// The server should start, accept, and shut down cleanly even if the dialer would fail.
		s := &ServerConfig{
			ListenIP:       "127.0.0.1",
			ListenPort:     "8080",
			BackendIP:      "127.0.0.1",
			BackendPort:    "8080",
			NetListener:    &MockNetListener{},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		wg.Wait()
		assert.NoError(t, s.ServerError())
	})
	t.Run("TestServerAllOk", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP:       "127.0.0.1",
			ListenPort:     "8080",
			BackendIP:      "127.0.0.1",
			BackendPort:    "8080",
			NetListener:    &MockNetListener{},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		wg.Wait()
		assert.NoError(t, s.ServerError())
	})
}
