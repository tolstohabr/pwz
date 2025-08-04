package mw

import (
	"context"

	"github.com/ulule/limiter/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func RateLimiterInterceptor(limiter *limiter.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sender := "unknown"
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if s, ok := md["sender"]; ok {
				sender = s[0]
			}
		}

		limiterCtx, err := limiter.Get(ctx, sender)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if limiterCtx.Reached {
			return nil, status.Error(codes.ResourceExhausted, "rate limited")
		}

		return handler(ctx, req)
	}
}
