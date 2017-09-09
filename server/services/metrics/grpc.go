package metrics

import (
	"context"

	"google.golang.org/grpc"
)

func (service *MetricsService) grpcRequestWrapper(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	service.aggregator.BeginRequest()
	res, err := handler(ctx, req)
	service.aggregator.EndRequest()

	return res, err
}
