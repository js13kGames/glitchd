package interfaces

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type HttpServerInterface struct {
	addrs   []string
	server  *http.Server
	servers []http.Server
	logger  *zap.Logger
}

func NewHttpServerInterface(addrs []string, cert *tls.Certificate, logger *zap.Logger) *HttpServerInterface {
	// @todo The recovery middleware, just as any other potential global middleware, should probably be moved
	// out and registered elsewhere.
	handler := gin.New()
	handler.Use(gin.Recovery())

	return &HttpServerInterface{
		addrs:  addrs,
		logger: logger,
		server: &http.Server{
			Handler: handler,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{*cert},
			},
		},
	}
}

func (iface *HttpServerInterface) GetKind() string {
	return "HTTP"
}

func (iface *HttpServerInterface) GetHandler() *gin.Engine {
	return iface.server.Handler.(*gin.Engine)
}

// Interface impl.
func (iface *HttpServerInterface) Start() {
	for _, addr := range iface.addrs {
		go func(addr string) {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				iface.logger.Fatal("Failed to bind", zap.Error(err))
			}

			if err := iface.server.ServeTLS(listener, "", ""); err != nil {
				iface.logger.Fatal("Failed to serve", zap.Error(err))
			}
		}(addr)
	}
}

// Interface impl.
func (iface *HttpServerInterface) Stop(deadline *time.Time) {
	var wg sync.WaitGroup
	wg.Add(len(iface.servers))

	if deadline == nil {
		iface.stopForce(&wg)
		// Stop() itself will most likely be called in a sub goroutine so we want to prevent
		// exit until all servers have finished their on-shutdown procedures, even if
		// it's a forceful close.
		wg.Wait()
		return
	}

	ctx, cancel := context.WithDeadline(context.Background(), *deadline)
	iface.stopGrace(ctx, &wg)
	wg.Wait()
	cancel()
}

// stopForce immediately closes all active servers grouped in this interface, effectively
// closing all opened listeners and dropping all peer connections.
func (iface *HttpServerInterface) stopForce(wg *sync.WaitGroup) {
	for _, server := range iface.servers {
		go func(server http.Server) {
			if err := server.Close(); err != nil {
				// Non-fatal.
				iface.logger.Error("Failed to close forcefully", zap.Error(err))
			}
			wg.Done()
		}(server)
	}
}

// stopGrace closes all active servers grouped in this interface while trying to allow all
// opened listeners and peer connections to gracefully finish their tasks up to a deadline
// defined by the passed in Context.
func (iface *HttpServerInterface) stopGrace(ctx context.Context, wg *sync.WaitGroup) {
	for _, server := range iface.servers {
		go func(server http.Server) {
			if err := server.Shutdown(ctx); err != nil {
				// Non-fatal.
				iface.logger.Error("Failed to close gracefully", zap.Error(err))
			}
			wg.Done()
		}(server)
	}
}
