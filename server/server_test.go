package server

import (
	"context"
	"testing"
	"sync"

	"github.com/stretchr/testify/assert"
)	

func TestServerConfig(t *testing.T) {
	t.Run("TestServerListenerError", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP: "867.053.009.000",
			ListenPort: "8080",
			BackendIP: "127.0.0.1",
			BackendPort: "8080",
			NetListener: &MockNetListener{SendError: true},
			Dialer: &MockNetDialer{},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		assert.Error(t, s.serverError)
	})
	t.Run("TestServerAcceptError", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP: "127.0.0.1",
			ListenPort: "8080",
			BackendIP: "127.0.0.1",
			BackendPort: "8080",
			NetListener: &MockNetListener{ListenerAcceptError: true},
			Dialer: &MockNetDialer{},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		assert.Error(t, s.serverError)
	})
	t.Run("TestServerDialError", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP: "127.0.0.1",
			ListenPort: "8080",
			BackendIP: "127.0.0.1",
			BackendPort: "8080",
			NetListener: &MockNetListener{},
			Dialer: &MockNetDialer{DialError: true},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		assert.Error(t, s.serverError)
	})
	t.Run("TestServerAllOk", func(t *testing.T) {
		// Create a new ServerConfig struct
		s := &ServerConfig{
			ListenIP: "127.0.0.1",
			ListenPort: "8080",
			BackendIP: "127.0.0.1",
			BackendPort: "8080",
			NetListener: &MockNetListener{},
			Dialer: &MockNetDialer{},
			HandlerFactory: &MockHandlerFactory{},
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s.StartServer(&wg, ctx)
		wg.Wait()
		assert.NoError(t, s.serverError)
	})
}