package main

import (
	"database/sql"
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

	"github.com/joho/godotenv"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	grpcAddress = "localhost:50051"
)

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Формируем DSN
	dsn := getPostgresDSN()
	fmt.Println("DSN:", dsn) // Для отладки

	if dsn == "" {
		log.Fatal("Postgres DSN is empty — проверь переменные окружения")
	}

	// Подключение к БД
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Инициализация сервиса
	storage := storage.NewPgStorage(db)
	orderService := service.NewOrderService(storage)
	orderServer := order.NewHandler(orderService)

	// Лимитер
	rate := limiter.Rate{Period: 10 * time.Second, Limit: 2}
	store := memory.NewStore()
	rateLimiter := mw.RateLimiterInterceptor(limiter.New(store, rate))

	// gRPC сервер
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mw.LoggingInterceptor,
			mw.ValidateInterceptor,
			rateLimiter,
		),
	)

	reflection.Register(grpcServer)
	desc.RegisterNotifierServer(grpcServer, orderServer)

	// Слушаем порт
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getPostgresDSN() string {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_BD")

	if user == "" || password == "" || host == "" || port == "" || dbname == "" {
		log.Println("нету")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}
