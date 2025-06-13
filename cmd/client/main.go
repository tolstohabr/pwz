package main

import (
	"context"
	"log"
	"time"

	desc "PWZ1.0/pkg/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	grpcAddress = "localhost:50051"
)

func main() {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to create gRPC client %v", err)
	}
	//defer conn.Close()
	///*
	//вместо defer conn.Close() препод вот это сделал и вышла какая-то ошибка Failed to close gRPC client <nil>
	//работает, но мне это не нравится
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close gRPC client %v", err)
		}
	}()
	//*/

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := desc.NewNotifierClient(conn)

	if err := sendMessage(ctx, client); err != nil {
		log.Fatalf("failed to send message %v", err)
	}
}

func sendMessage(ctx context.Context, client desc.NotifierClient) error {
	//тут задаем поверх контекста метаданные
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.MessageRequest{
		Text: "FIRST MESSAGE from client",
	}

	res, err := client.SendMessage(ctx, req)
	if err != nil {
		return err
	}

	log.Printf("response from server: %v", res)
	return nil
}
