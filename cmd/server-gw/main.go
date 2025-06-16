package main

import (
	"context"
	"log"
	"net/http"

	desc "PWZ1.0/pkg/pwz"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcAddress = "localhost:50051"
	httpAddress = "localhost:50052"
)

func main() {
	ctx := context.Background()
	mux := runtime.NewServeMux()
	err := desc.RegisterNotifierHandlerFromEndpoint(ctx, mux, grpcAddress, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	})
	if err != nil {
		log.Fatalf("RegisterNotifierHandlerFromEndpoint err: %v", err)
	}

	log.Printf("http server running on %v", httpAddress)
	if err := http.ListenAndServe(httpAddress, mux); err != nil {
		log.Fatalf("http server running err: %v", err)
	}
}
