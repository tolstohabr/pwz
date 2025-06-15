package main

import (
	"context"
	"fmt"
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

	if err := acceptOrder(ctx, client, 103, 1, time.Now().Add(24*time.Hour), ptr(desc.PackageType_PACKAGE_TYPE_BOX), 1.0, 100.0); err != nil {
		log.Fatalf("failed to accept order: %v", err)
	}

	/*if err := listOrders(ctx, client, 1, true, nil, 0, 15); err != nil {
		log.Fatalf("failed to list orders: %v", err)
	}*/
}

func acceptOrder(ctx context.Context, client desc.NotifierClient, orderID uint64, userID uint64, expiresAt time.Time, pkg *desc.PackageType, weight float32, price float32) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.AcceptOrderRequest{
		OrderId:   orderID,
		UserId:    userID,
		ExpiresAt: timestamppb.New(expiresAt),
		Package:   pkg,
		Weight:    weight,
		Price:     price,
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

func listOrders(ctx context.Context, client desc.NotifierClient, userID uint64, inPvz bool, lastN *uint32, page uint32, limit uint32) error {
	req := &desc.ListOrdersRequest{
		UserId: userID,
		InPvz:  inPvz,
		LastN:  lastN,
	}

	if page > 0 || limit > 0 {
		req.Pagination = &desc.Pagination{
			Page:        page,
			CountOnPage: limit,
		}
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "bearer my-token"))

	resp, err := client.ListOrders(ctx, req)
	if err != nil {
		return fmt.Errorf("ListOrders failed: %w", err)
	}

	for _, order := range resp.Orders {
		fmt.Printf("Order ID: %d, User ID: %d, Status: %s, Expires: %v, Weight: %.2f, Price: %.2f",
			order.OrderId,
			order.UserId,
			order.Status.String(),
			order.ExpiresAt.AsTime().Format(time.RFC3339),
			order.Weight,
			order.TotalPrice)

		if order.Package != nil {
			fmt.Printf(", Package: %s\n", order.Package.String())
		} else {
			fmt.Println(", Package: none")
		}
	}
	fmt.Printf("Total orders: %d\n", resp.Total)

	return nil
}
