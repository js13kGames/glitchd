package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/js13kgames/glitchd/server/services/items/types"
)

// @todo The GET/PUT req/res cycle is suboptimal under the hood currently. Put Messages get marshalled
// in the client, unmarshalled on the server, then stored. Get Messages get marshalled on the server,
// and unmarshalled in the client. Ideally we could skip the Put unmarshalling and Get marshalling on the server,
// limit messages to just the raw payload (passing keys as metadata instead), and store the already marshalled
// payloads directly, to skip double encoding.
// Note that this would require a custom codec and ideally just for the ItemsStore, so the protos would need
// to be split accordingly and the service would likely not be able to use most of the auto generated defs.
type Service struct{}

//
//
//
func (s *Service) Get(ctx context.Context, in *StoreGetRequest) (*StoreGetResponse, error) {
	val, err := ctx.Value(storeCtxKey).(*types.Store).Get(in.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve the item.")
	}

	if val == nil {
		return nil, status.Errorf(codes.NotFound, "No value found for the requested key.")
	}

	return &StoreGetResponse{Value: val}, nil
}

//
//
//
func (s *Service) Put(ctx context.Context, in *StorePutRequest) (*Empty, error) {
	// Theoretically a put with a nil slice is equal to deleting an item, but we do want to ensure
	// the calls are precise in intent and separate in concerns.
	if in.Value == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot put empty values. Call delete instead if you intended to delete an item.")
	}

	if err := ctx.Value(storeCtxKey).(*types.Store).Put(in.Key, in.Value); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to store the item.")
	}

	return &Empty{}, nil
}

//
//
//
func (s *Service) Delete(ctx context.Context, in *StoreDeleteRequest) (*Empty, error) {
	if err := ctx.Value(storeCtxKey).(*types.Store).Delete(in.Key); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete the item.")
	}

	return &Empty{}, nil
}
