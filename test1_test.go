package homework

import (
	"context"
	"log"
	"testing"

	desc "PWZ1.0/pkg/pwz"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	grpcAddress    = "localhost:50051"
	DateTimeFormat = "2006-01-02 15:04:05"
)

func TestBanner(t *testing.T) {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := desc.NewNotifierClient(conn)
	ctx := metadata.AppendToOutgoingContext(context.Background(), "sender", "go-client", "client-version", "1.0")

	t.Run("get order history", func(t *testing.T) {
		req := &desc.OrderHistoryRequest{
			OrderId: 20012,
		}

		res, err := client.GetOrderHistory(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, res.GetHistory(), 3)
	})

	t.Run("get history", func(t *testing.T) {
		req := &desc.GetHistoryRequest{
			Pagination: &desc.Pagination{
				Page:        0,
				CountOnPage: 10,
			},
		}

		res, err := client.GetHistory(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, res.GetHistory(), 10)
	})
}
