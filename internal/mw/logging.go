package mw

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	log.Printf("SendMessage: %v", req)
	resp, err = handler(ctx, req)
	log.Printf("Error: %v", err)

	return resp, err
}
