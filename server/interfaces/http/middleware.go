package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func BearerTokenInterceptor(ctx *gin.Context) {
	if len(ctx.Request.Header["Authorization"]) == 0 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Only take the first Authorization header present into account.
	auth := ctx.Request.Header["Authorization"][0]

	if len(auth) <= 7 || strings.ToLower(auth[0:7]) != "bearer " {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Note: Not checking validity here since it will depend on the context the token
	// is being used, which we will determine further down the line.
	ctx.Set("token", auth[7:])
}

func PrivilegedTokenVerifier(key string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, ok := ctx.Keys["token"].(string)
		if !ok {
			// We are panicking here because it's a program logic error if it happens - it would
			// signify that either the token was not intercepted at all or that it was not present
			// but the request was not aborted regardless.
			panic("Cannot verify privileged key - 'token' is not set in the request context.")
		}
		if token != key {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}
