package metrics

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/js13kgames/glitchd/server/interfaces/http"
)

//
//
//
func (service *MetricsService) httpRequestWrapper(ctx *gin.Context) {
	service.aggregator.BeginRequest()
	ctx.Next()
	service.aggregator.EndRequest()
}

//
//
//
func (service *MetricsService) registerHttpMiddleware(router *gin.Engine) {
	router.Use(gin.LoggerWithWriter(os.Stdout))
	router.Use(service.httpRequestWrapper)
}

//
//
//
func (service *MetricsService) registerHttpRoutes(router *gin.Engine) {
	router.GET("/metrics", http.BearerTokenInterceptor, http.PrivilegedTokenVerifier(service.key), func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(c.Writer).Encode(service.aggregator.Collect()); err != nil {
			c.AbortWithError(500, err)
		}
	})
}
