package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	grpcAddress    = "localhost:50051"
	DateTimeFormat = "2006-01-02 15:04:05"
)

type ImportJSON struct {
	Orders []struct {
		OrderID   uint64  `json:"order_id"`
		UserID    uint64  `json:"user_id"`
		ExpiresAt string  `json:"expires_at"`
		Package   string  `json:"package"`
		Weight    float32 `json:"weight"`
		Price     float32 `json:"price"`
	} `json:"orders"`
}

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

	if err := listReturns(ctx, client, 0, 10); err != nil {
		log.Fatalf("failed to list returns: %v", err)
	}
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

func processOrders(ctx context.Context, client desc.NotifierClient, userID uint64, action desc.ActionType, orderIDs []uint64) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.ProcessOrdersRequest{
		UserId:   userID,
		Action:   action,
		OrderIds: orderIDs,
	}

	res, err := client.ProcessOrders(ctx, req)
	if err != nil {
		return err
	}

	fmt.Println("Processed orders:")
	for _, id := range res.Processed {
		fmt.Printf("- %d\n", id)
	}

	if len(res.Errors) > 0 {
		fmt.Println("Failed orders:")
		for _, id := range res.Errors {
			fmt.Printf("- %d\n", id)
		}
	}

	return nil
}

func returnOrder(ctx context.Context, client desc.NotifierClient, orderID uint64) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.OrderIdRequest{
		OrderId: orderID,
	}

	res, err := client.ReturnOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("ReturnOrder failed: %w", err)
	}

	fmt.Printf("Order returned to courier: ID=%d, Status=%s\n", res.OrderId, res.Status.String())
	return nil
}

func listReturns(ctx context.Context, client desc.NotifierClient, page uint32, limit uint32) error {
	req := &desc.ListReturnsRequest{
		Pagination: &desc.Pagination{
			Page:        page,
			CountOnPage: limit,
		},
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	resp, err := client.ListReturns(ctx, req)
	if err != nil {
		return fmt.Errorf("ListReturns failed: %w", err)
	}

	fmt.Println("Returns list:")
	for _, ret := range resp.Returns {
		fmt.Printf("Order ID: %d, User ID: %d, Status: %s, Expires: %v, Weight: %.2f, Price: %.2f",
			ret.OrderId,
			ret.UserId,
			ret.Status.String(),
			ret.ExpiresAt.AsTime().Format(time.RFC3339),
			ret.Weight,
			ret.TotalPrice)

		if ret.Package != nil {
			fmt.Printf(", Package: %s\n", ret.Package.String())
		} else {
			fmt.Println(", Package: none")
		}
	}

	fmt.Printf("PAGE: %d LIMIT: %d\n", page, limit)
	return nil
}

func getHistory(ctx context.Context, client desc.NotifierClient, page, count uint32) error {
	req := &desc.GetHistoryRequest{
		Pagination: &desc.Pagination{
			Page:        page,
			CountOnPage: count,
		},
	}

	resp, err := client.GetHistory(ctx, req)
	if err != nil {
		return fmt.Errorf("GetHistory failed: %w", err)
	}

	fmt.Println("Order History:")
	for _, h := range resp.GetHistory() {
		createdAt := h.GetCreatedAt().AsTime().Format(DateTimeFormat)
		fmt.Printf("OrderID: %d, Status: %s, CreatedAt: %s\n",
			h.GetOrderId(), h.GetStatus().String(), createdAt)
	}

	return nil
}

func importOrdersFromFile(ctx context.Context, client desc.NotifierClient, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var parsed ImportJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	var orders []*desc.AcceptOrderRequest
	for _, o := range parsed.Orders {
		t, err := time.Parse(time.RFC3339, o.ExpiresAt)
		if err != nil {
			return fmt.Errorf("invalid time format for order %d: %w", o.OrderID, err)
		}

		var pkg desc.PackageType = desc.PackageType_PACKAGE_TYPE_UNSPECIFIED
		if o.Package != "" {
			if parsedPkg, ok := desc.PackageType_value[o.Package]; ok {
				pkg = desc.PackageType(parsedPkg)
			} else {
				log.Printf("Warning: unknown package type for order %d: %s", o.OrderID, o.Package)
			}
		}

		orders = append(orders, &desc.AcceptOrderRequest{
			OrderId:   o.OrderID,
			UserId:    o.UserID,
			ExpiresAt: timestamppb.New(t),
			Package:   &pkg,
			Weight:    o.Weight,
			Price:     o.Price,
		})
	}

	req := &desc.ImportOrdersRequest{
		Orders: orders,
	}

	resp, err := client.ImportOrders(ctx, req)
	if err != nil {
		return fmt.Errorf("ImportOrders RPC failed: %w", err)
	}

	fmt.Printf("Импортировано заказов: %d\n", resp.Imported)
	if len(resp.Errors) > 0 {
		fmt.Println("Ошибки при импорте (order_ids):")
		for _, id := range resp.Errors {
			fmt.Printf("- %d\n", id)
		}
	}
	return nil
}

func getOrderHistory(ctx context.Context, client desc.NotifierClient, orderID uint64) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	req := &desc.OrderHistoryRequest{
		OrderId: orderID,
	}

	res, err := client.GetOrderHistory(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get order history: %w", err)
	}

	fmt.Printf("История заказа %d:\n", orderID)
	for _, h := range res.History {
		fmt.Printf("%s Время: %s\n", h.Status.String(), h.CreatedAt.AsTime().Format(DateTimeFormat))
	}
	return nil
}
