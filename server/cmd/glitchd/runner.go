package main

import (
	"crypto/tls"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/js13kgames/glitchd/server"
	"github.com/js13kgames/glitchd/server/interfaces"
	"github.com/js13kgames/glitchd/server/metrics"
	"github.com/js13kgames/glitchd/server/services"
	"github.com/js13kgames/glitchd/server/services/items"
	metricsSrv "github.com/js13kgames/glitchd/server/services/metrics"
)

type Runner struct {
	closing chan struct{}
	Closed  chan struct{}
	logger  *zap.Logger
	manager *services.Manager
}

func (runner *Runner) Run() {
	var (
		restKey  string
		restAddr string
		rpcAddr  string
		dbFile   string
	)

	restKey = os.Getenv("GLITCHD_REST_KEY")
	if len(restKey) == 0 {
		runner.logger.Fatal("Failed to initialize: no non-empty GLITCHD_REST_KEY envvar present")
	}

	restAddr = os.Getenv("GLITCHD_REST_ADDRESS")
	if len(restAddr) == 0 {
		restAddr = ":13313"
	}

	rpcAddr = os.Getenv("GLITCHD_RPC_ADDRESS")
	if len(rpcAddr) == 0 {
		rpcAddr = ":13312"
	}

	dbFile = os.Getenv("GLITCHD_DB")
	if len(dbFile) == 0 {
		dbFile = "glitchd.db"
	}

	certificate := runner.loadServerCertificate()

	// Write the PID or just log the PID. In any case of failure don't stop processing however -
	// simply log a warning instead.
	if err := runner.handleWritePidFile(); err != nil {
		runner.logger.Warn("Failed to write the PID file", zap.Error(err))
	}

	router := gin.New()
	router.Use(gin.Recovery())

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		runner.logger.Fatal(err.Error())
	}
	defer db.Close()

	// @todo Both server interfaces and services should be fully configurable (ideally services would
	// simply define a hard or soft dependency on a particular interface and we'd infer what and how to load
	// based on that).
	runner.logger.Debug("Registering services")
	runner.manager = services.NewServiceManager(runner.logger,
		[]server.Interface{
			interfaces.NewGrpcServerInterface([]string{rpcAddr}, certificate, runner.logger),
			interfaces.NewHttpServerInterface([]string{restAddr}, certificate, runner.logger),
		},
		[]services.Service{
			metricsSrv.NewMetricsService(metrics.NewGlobalAggregator(), restKey),
			items.NewItemsService(db, []byte("stores"), restKey, runner.logger),
		})

	runner.logger.Debug("Bootstrapping services")
	runner.manager.Bootstrap()

	runner.logger.Debug("Running service manager")
	go runner.manager.Run()

	runner.waitForSignals()
}

//
//
//
func (runner *Runner) Close(gracePeriod time.Duration) {
	close(runner.closing)

	if runner.manager != nil {
		runner.manager.Stop(&gracePeriod)
	}

	close(runner.Closed)
}

//
//
// @todo This would make much more sense on a per-interface/per-service basis.
func (runner *Runner) loadServerCertificate() *tls.Certificate {
	var (
		certFile string
		keyFile  string
	)

	certFile = os.Getenv("GLITCHD_SERVER_CERT")
	if len(certFile) == 0 {
		certFile = "server.crt"
	}

	keyFile = os.Getenv("GLITCHD_SERVER_KEY")
	if len(keyFile) == 0 {
		keyFile = "server.key"
	}

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		runner.logger.Fatal("Failed to load server key pair",
			zap.Error(err),
			zap.String("certFile", certFile),
			zap.String("keyFile", keyFile),
		)
		// Not executed. logger.Fatal is an OS exit, but the compiler would complain otherwise.
		return nil
	}

	return &certificate
}

//
//
//
func (runner *Runner) handleWritePidFile() error {
	runner.logger.Info("Process started", zap.Int("PID", os.Getpid()))

	// @todo Write the PID - once we have the path configurable.

	return nil
}

//
//
//
func (runner *Runner) waitForSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	runner.logger.Debug("Waiting for system signals")

	<-signals

	gracePeriod := time.Duration(time.Second * 30)

	runner.logger.Info("System signal received, initializing shutdown...", zap.Duration("grace", gracePeriod))

	// Perform the shutdown in its own goroutine since we want to remain responsive to additional signals (and timeouts)
	// while it's going on. Most children will spawn their own goroutines for shutdown on top of this, several in some
	// cases, so this will need monitoring once complexity grows - primarily for races to the underlying storage.
	go runner.Close(gracePeriod)

	select {
	case <-signals:
		runner.logger.Info("Second signal received, doing hard shutdown")
	case <-time.After(gracePeriod):
		runner.logger.Info("Shutdown timeout exceeded, doing hard shutdown")
	case <-runner.Closed:
		runner.logger.Info("Service shutdown complete")
	}
}
