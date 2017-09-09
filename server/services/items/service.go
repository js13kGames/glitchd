package items

import (
	"go.uber.org/zap"

	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/js13kgames/glitchd/server"
	"github.com/js13kgames/glitchd/server/interfaces"
	"github.com/js13kgames/glitchd/server/services"
	grpcService "github.com/js13kgames/glitchd/server/services/items/grpc"
	restService "github.com/js13kgames/glitchd/server/services/items/rest"
	"github.com/js13kgames/glitchd/server/services/items/types"
	metricsService "github.com/js13kgames/glitchd/server/services/metrics"
)

type ItemsService struct {
	logger  *zap.Logger
	restKey string
	stores  *types.StoreRepository
}

func NewItemsService(db *bolt.DB, bucketKey []byte, key string, logger *zap.Logger) *ItemsService {
	// @todo Validate the params - once we have a proper config pipeline in place.
	stores, err := types.LoadStoreRepository(db, bucketKey)
	if err != nil {
		logger.Fatal(err.Error())
	}

	return &ItemsService{
		logger:  logger,
		restKey: key,
		stores:  stores,
	}
}

func (service *ItemsService) GetName() string {
	return "items"
}

func (service *ItemsService) Bootstrap(manager *services.Manager, ifaces []server.Interface, srvcs []services.Service) {

	var httpHandlers []*gin.Engine

	for _, iface := range ifaces {
		switch v := iface.(type) {

		// @todo While we currently retain full control and knowledge of what's being
		// initialized, once proper configs are in place and/or services become pluggable,
		// this will also need a means of filtering through interfaces on other criteria
		// than just their type.
		case *interfaces.GrpcServerInterface:
			grpcService.RegisterStoreServer(v.GetServer(), &grpcService.Service{})
			v.PushUnaryInterceptor(grpcService.UnaryStoreExtractor(service.stores))

		case *interfaces.HttpServerInterface:
			httpHandlers = append(httpHandlers, v.GetHandler())
			restService.RegisterBaseRoutes(v.GetHandler(), service.restKey, service.stores)
		}
	}

	for _, srvc := range srvcs {
		if _, ok := srvc.(*metricsService.MetricsService); ok {
			for _, handler := range httpHandlers {
				restService.RegisterMetricsRoutes(handler, service.restKey, service.stores)
				service.stores.RegisterMetricsTicks(manager)
			}
			// Can't imagine a reason for there being more than one Metrics service registered
			// at runtime.
			break
		}
	}
}

func (service *ItemsService) Start() {
	// No-op - we only register with global interfaces.
}

func (service *ItemsService) Stop(deadline *time.Time) {
	// No-op - we only register with global interfaces.
}
