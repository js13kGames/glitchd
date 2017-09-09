package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/js13kgames/glitchd/server/interfaces/http"
	"github.com/js13kgames/glitchd/server/services/items/types"
)

func RegisterBaseRoutes(router *gin.Engine, key string, storeRepository *types.StoreRepository) {
	stores := router.Group("/stores", http.BearerTokenInterceptor, http.PrivilegedTokenVerifier(key))
	stores.GET("", storesListHandler(storeRepository))
	stores.POST("", storesInsertHandler(storeRepository))

	{
		store := stores.Group("/:storeId", storeFromParamMapper(storeRepository))
		store.PATCH("", storesStorePatchHandler(storeRepository))
		store.DELETE("", storesStoreDeleteHandler(storeRepository))

		store.POST("/token", storesStoreTokenRotateHandler(storeRepository))
	}
}

func RegisterMetricsRoutes(router *gin.Engine, key string, storeRepository *types.StoreRepository) {
	router.GET("/stores/:storeId/metrics",
		http.BearerTokenInterceptor,
		http.PrivilegedTokenVerifier(key),
		storeFromParamMapper(storeRepository),
		storesStoreMetricsHandler(),
	)
}
