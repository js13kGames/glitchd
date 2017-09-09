package interfaces

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//
//
//
type GrpcServerInterface struct {
	isClosing         *uint32
	addrs             []string
	server            *grpc.Server
	logger            *zap.Logger
	interceptorsUnary []grpc.UnaryServerInterceptor
}

//
//
//
func NewGrpcServerInterface(addrs []string, cert *tls.Certificate, logger *zap.Logger) *GrpcServerInterface {
	iface := &GrpcServerInterface{
		isClosing: new(uint32),
		addrs:     addrs,
		logger:    logger,
	}

	iface.server = grpc.NewServer(
		grpc.Creds(credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.NoClientCert,
			Certificates: []tls.Certificate{*cert},
		})),

		// @todo Separate on a per-service basis.
		grpc.MaxRecvMsgSize(32*1024),
		grpc.UnaryInterceptor(iface.interceptUnary),
	)

	return iface
}

// Interface impl.
func (iface *GrpcServerInterface) GetKind() string {
	return "gRPC"
}

//
//
//
func (iface *GrpcServerInterface) GetServer() *grpc.Server {
	return iface.server
}

// Note: *Not* thread safe. Interceptor chains run on the *unsafe* assumption that the backing array
// only gets mutated during construction/initialization, and in sequence at that.
func (iface *GrpcServerInterface) PushUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) {
	iface.interceptorsUnary = append(iface.interceptorsUnary, interceptor)
}

func (iface *GrpcServerInterface) PrependUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) {
	iface.interceptorsUnary = append([]grpc.UnaryServerInterceptor{interceptor}, iface.interceptorsUnary...)
}

// Interface impl.
func (iface *GrpcServerInterface) Start() {
	// Currently we can get by without reflections. And if we do need them, this needs
	// to be moved out of serve into a bootstrap procedure.
	// reflection.Register(iface.server)

	// @fixme Should be set in Stop() instead, once the underlying server stops. However,
	// without proper sync the stop and the resetting of isClosing usually happens before
	// the listener reports ErrNetClosing which we want to catch, so this is just a poor
	// stopgap measure.
	atomic.StoreUint32(iface.isClosing, 0)

	// Bind to each configured address and start serving incoming connections.
	for _, addr := range iface.addrs {
		go func(addr string) {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				iface.logger.Fatal("Failed to bind", zap.Error(err))
			}

			if err := iface.server.Serve(listener); err != nil {
				// @todo There's an issue with shutdowns (either in std or in gRPC) - which I can't yet pinpoint -
				// which causes Serve (the underlying TCP listener to be more precise) to return
				// "use of closed network connection" due to an op on a closed listener socket.
				// This happens after a call to our Stop() which then stops grpc.Server, but the only actual
				// loop operating on the listener is the one started by grpc.Server.Serve() - eg. the issue
				// is external to glitchd.
				// It appears harmless - so we're doing the thing you should never do with errors, which is
				// silencing it. *If* it actually happens during a shutdown procedure, that is.
				// Also, ErrNetClosing is in the "poll" package, which is internal, so we're checking
				// for a hardcoded string...
				if atomic.LoadUint32(iface.isClosing) == 1 && strings.Contains(err.Error(), "use of closed network connection") {
					return
				}

				iface.logger.Fatal("Failed to serve", zap.Error(err))
			}
		}(addr)
	}
}

// Interface impl.
// @todo gRPC does not support deadlines on its own. We could wrap some timers given lots of spare time,
// but otherwise the main Runner will take care of imposing a hard deadline on the whole process.
func (iface *GrpcServerInterface) Stop(deadline *time.Time) {
	atomic.StoreUint32(iface.isClosing, 1)

	if deadline == nil {
		iface.server.Stop()
		return
	}

	iface.server.GracefulStop()
}

//
//
//
func (iface *GrpcServerInterface) interceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// Micro-optimization note: len() on a global/shared var is not cached as it is for local countables,
	// so doing one len() call in this case is ~10-20ns/op faster (and obviously scales with the number
	// of uncached len calls)
	n := len(iface.interceptorsUnary)

	if n > 1 {
		i := 0
		x := n - 1
		// Note: The var has to be declared first as it is called from within the lambda.
		// While the below works, relying on type inference (eg. chainHandler := ) without declaring
		// the var beforehand does not. Which on its own wasn't entirely unexpected, but if that's
		// the case, then the fact the below assignment works is unexpected.
		var chainHandler grpc.UnaryHandler
		chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
			// If it's the last interceptor, call the actual handler.
			if i == x {
				return handler(currentCtx, currentReq)
			}
			i++
			return iface.interceptorsUnary[i](currentCtx, currentReq, info, chainHandler)
		}

		return iface.interceptorsUnary[0](ctx, req, info, chainHandler)
	}

	if n == 1 {
		return iface.interceptorsUnary[0](ctx, req, info, handler)
	}

	// n == 0
	return handler(ctx, req)
}
