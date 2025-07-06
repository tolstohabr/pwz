package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"PWZ1.0/internal/app/order"
	"PWZ1.0/internal/mw"
	"PWZ1.0/internal/service"
	"PWZ1.0/internal/storage"
	desc "PWZ1.0/pkg/pwz"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const grpcAddress = "localhost:50051"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dsn := os.Getenv("DB_DSN")
	fmt.Println("DSN:", dsn)
	if dsn == "" {
		log.Fatal("Postgres DSN is empty")
	}

	//todo: увеличиваю время
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create pgxpool: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	storage := storage.NewPgStorage(db)
	orderService := service.NewOrderService(storage)
	orderServer := order.NewHandler(orderService)

	//todo: увеличиваю лимит
	rate := limiter.Rate{Period: 10 * time.Second, Limit: 100}
	store := memory.NewStore()
	rateLimiter := mw.RateLimiterInterceptor(limiter.New(store, rate))

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mw.LoggingInterceptor,
			mw.ValidateInterceptor,
			rateLimiter,
		),
	)

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
