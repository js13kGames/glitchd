package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/js13kgames/glitchd/server/services/items/types"
)

//
//
//
func storesListHandler(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, stores)
	}
}

//
//
//
func storesInsertHandler(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var store *types.Store

		if err := ctx.BindJSON(&store); err != nil {
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		store, err := stores.Create(store.Id, store.Token)
		if err != nil {
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.JSON(http.StatusOK, store)
	}
}

//
//
//
func storesStorePatchHandler(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			source = ctx.Keys["store"].(*types.Store)
			target *types.Store
		)

		if err := ctx.BindJSON(&target); err != nil {
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if target.OwnerId != 0 {
			source.OwnerId = target.OwnerId
		}

		if target.SubmissionId != 0 {
			source.SubmissionId = target.SubmissionId
		}

		// A bit of special treatment for manual Token changes (even though we don't expect those to happen,
		// the ability will be left in, in case a (temporary) lockout without purging the whole Store
		// is necessary.
		if target.Token != "" && target.Token != source.Token {
			delete(stores.Items, source.Token)
			source.Token = target.Token
		}

		stores.Save(source)

		ctx.Writer.WriteHeader(http.StatusNoContent)
	}
}

//
//
//
func storesStoreDeleteHandler(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		stores.Delete(ctx.Keys["store"].(*types.Store))
		ctx.Writer.WriteHeader(http.StatusNoContent)
	}
}

//
//
//
func storesStoreTokenRotateHandler(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resource := ctx.Keys["store"].(*types.Store)
		stores.RotateToken(resource)
		ctx.JSON(http.StatusOK, resource.Token)
	}
}

//
//
//
func storesStoreMetricsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resource := ctx.Keys["store"].(*types.Store)

		resource.Put("foo1", []byte("bar1ą"))
		resource.Put("foo2", []byte("bar2ą"))
		resource.Put("foo3", []byte("bar3ą"))
		resource.Delete("foo3")

		if metrics := resource.Metrics(); metrics != nil {
			ctx.JSON(http.StatusOK, metrics)
			return
		}

		ctx.Writer.WriteHeader(http.StatusNotFound)
	}
}
