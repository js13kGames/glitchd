package metrics

import (
	"time"

	"google.golang.org/grpc"

	"github.com/js13kgames/glitchd/server"
	"github.com/js13kgames/glitchd/server/interfaces"
	"github.com/js13kgames/glitchd/server/metrics"
	"github.com/js13kgames/glitchd/server/services"
)

type MetricsService struct {
	key        string
	aggregator *metrics.GlobalAggregator
}

func NewMetricsService(aggregator *metrics.GlobalAggregator, key string) *MetricsService {
	return &MetricsService{
		aggregator: aggregator,
		key:        key,
	}
}

//
func (service *MetricsService) GetName() string {
	return "metrics"
}

//
func (service *MetricsService) Bootstrap(manager *services.Manager, ifaces []server.Interface, srvcs []services.Service) {
	service.aggregator.Bootstrap(manager)

	for _, iface := range ifaces {
		switch v := iface.(type) {
		case *interfaces.HttpServerInterface:
			service.registerHttpMiddleware(v.GetHandler())
			service.registerHttpRoutes(v.GetHandler())

		case *interfaces.GrpcServerInterface:
			v.PrependUnaryInterceptor(grpc.UnaryServerInterceptor(service.grpcRequestWrapper))
		}
	}
}

//
func (service *MetricsService) Start() {
	// No-op - we only register with global interfaces.
}

//
func (service *MetricsService) Stop(deadline *time.Time) {
	// No-op - we only register with global interfaces.
}
