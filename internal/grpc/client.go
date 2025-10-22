package grpc

import (
	"context"
	"fmt"
	"time"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	webikev1 "github.com/sm8ta/grpc_webike/gen/go/webike"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type BikeClient struct {
	api webikev1.BikeServiceClient
	log ports.LoggerPort
}

func NewBikeClient(
	ctx context.Context,
	log ports.LoggerPort,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*BikeClient, error) {
	const op = "grpc.NewBikeClient"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := webikev1.NewBikeServiceClient(cc)

	log.Info("Connected to Bike service", map[string]interface{}{
		"addr": addr,
	})

	return &BikeClient{
		api: grpcClient,
		log: log,
	}, nil
}

func (c *BikeClient) GetBikes(ctx context.Context, userID string) (*webikev1.GetBikesResponse, error) {
	const op = "BikeClient.GetBikes"

	c.log.Debug("Calling Bike service", map[string]interface{}{
		"user_id": userID,
	})

	resp, err := c.api.GetBikes(ctx, &webikev1.GetBikesRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c.log.Debug("Received bikes from Bike service", map[string]interface{}{
		"user_id":     userID,
		"bikes_count": len(resp.Bikes),
	})

	return resp, nil
}

// InterceptorLogger adapts ports.LoggerPort to interceptor logger.
func InterceptorLogger(l ports.LoggerPort) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		// Fields to map convert
		fieldsMap := make(map[string]interface{})
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fmt.Sprintf("%v", fields[i])
				fieldsMap[key] = fields[i+1]
			}
		}

		// Usig gRPC methods with context
		switch lvl {
		case grpclog.LevelInfo:
			l.InfoGRPC(ctx, msg, fieldsMap)
		case grpclog.LevelWarn:
			l.WarnGRPC(ctx, msg, fieldsMap)
		case grpclog.LevelError:
			l.ErrorGRPC(ctx, msg, fieldsMap)
		default:
			l.DebugGRPC(ctx, msg, fieldsMap)
		}
	})
}
