package main

import (
	"context"
	"log"
	"time"

	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	grpcAddress = "localhost:50051"
)

func main() {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create gRPC client %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close gRPC client %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := desc.NewNotifierClient(conn)

	if err := acceptOrder(ctx, client); err != nil {
		log.Fatalf("failed to accept pwz: %v", err)
	}
}

// TODO: accept order
func acceptOrder(ctx context.Context, client desc.NotifierClient) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.AcceptOrderRequest{
		OrderId:   3000,
		UserId:    1,
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		Package:   ptr(desc.PackageType_PACKAGE_TYPE_UNSPECIFIED),
		Weight:    1,
		Price:     100,
	}

	res, err := client.AcceptOrder(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("Order accepted: %v", res)
	return nil
}

func ptr[T any](v T) *T {
	return &v
}
