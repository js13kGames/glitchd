package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/js13kgames/glitchd/server/services/items/types"
)

type ctxKey uint8

const storeCtxKey ctxKey = 0

func UnaryStoreExtractor(stores *types.StoreRepository) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			md metadata.MD
			ok bool
		)

		md, ok = metadata.FromIncomingContext(ctx)

		// Expecting metadata to be always present and at the very least contain exactly one token. No more, no less.
		if !ok || len(md["token"]) != 1 || len(md["token"][0]) != types.TOKEN_LENGTH {
			return nil, status.Errorf(codes.Unauthenticated, "Missing access token.")
		}

		// Note: Returning 403 instead of 404 here because a Store must always be present for a valid token.
		// No store mapped to the given token effectively means the token is invalid.
		store := stores.Items[md["token"][0]]
		if store == nil {
			return nil, status.Errorf(codes.PermissionDenied, "Unknown access token.")
		}

		return handler(context.WithValue(ctx, storeCtxKey, store), req)
	}
}
