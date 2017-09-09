package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/js13kgames/glitchd/server/services/items/types"
)

// Note: Currently not used since we decided to move the tenant facing API via gRPC instead.
func storeFromTokenMapper(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Keys["token"].(string)
		if token == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Note: Returning 403 instead of 404 here because a Store must always be present for a valid token.
		// No store mapped to the given token effectively means the token is invalid.
		store := stores.Items[token]
		if store == nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.Set("store", store)
	}
}

func storeFromParamMapper(stores *types.StoreRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Note: Hardcoded assumption that the route layout does not change and the param remains
		// the first in order for the /stores family of endpoints.
		id, err := strconv.ParseUint(ctx.Params[0].Value, 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		store := stores.GetById(uint16(id))
		if store == nil {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set("store", store)
	}
}
