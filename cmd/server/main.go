package main

import (
	"log"
	"net"

	"PWZ1.0/internal/app/order"
	"PWZ1.0/internal/service"
	"PWZ1.0/internal/storage"
	desc "PWZ1.0/pkg/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcAddress = "localhost:50051"
)

func main() {

	storage := storage.NewFileStorage("orders.json")
	orderService := service.NewOrderService(storage)
	orderServer := order.NewHandler(orderService)

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	desc.RegisterNotifierServer(grpcServer, orderServer)

	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
