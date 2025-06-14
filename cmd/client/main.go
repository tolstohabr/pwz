package main

import (
	"context"
	"log"
	"time"

	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
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

	if err := sendMessage(ctx, client); err != nil {
		log.Fatalf("failed to send message %v", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client = desc.NewNotifierClient(conn)

	if err := acceptOrder(ctx, client); err != nil {
		log.Fatalf("failed to accept pwz: %v", err)
	}
}

func sendMessage(ctx context.Context, client desc.NotifierClient) error {
	//тут задаем поверх контекста метаданные
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	comment := "this is the comment from gekon"
	req := &desc.MessageRequest{
		Text:     "low priority",
		Priority: desc.Priority_PRIORITY_LOW,
		Delay:    &durationpb.Duration{Seconds: 2},
		Tags:     []string{"skull", "butterfly", "innocent"},
		Comment:  &comment,
		Title:    "gekonito bombito",
	}
	if err := req.ValidateAll(); err != nil {
		return err
	}

	res, err := client.SendMessage(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("response from server: %v", res)
	return nil
}

func acceptOrder(ctx context.Context, client desc.NotifierClient) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.AcceptOrderRequest{
		OrderId:   123456,
		UserId:    456,
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		Package:   ptr(desc.PackageType_PACKAGE_TYPE_UNSPECIFIED),
		Weight:    1.5,
		Price:     100.0,
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
