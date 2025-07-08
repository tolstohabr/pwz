package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

func main() {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := desc.NewNotifierClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// слэйв
	if err := listReturns(ctx, client, 0, 2); err != nil {
		log.Fatalf("failed to list returns: %v", err)
	}
	// мастер
	orderIDs := []uint64{30001, 30002, 30003, 30004, 30005, 30006, 30007, 30008, 30009, 30010, 30011, 30012, 30013,
		30014, 30015, 30016, 30017, 30019, 30020, 30021, 30022, 30023, 30025, 30026, 30027, 30028, 30029, 30030,
		30031, 30032, 30033, 30034, 30035, 30036, 30037, 30038, 30039, 30040, 30041, 30042, 30043, 30044, 30045}
	batchSize := 2

	var jobs []processJob
	for i := 0; i < len(orderIDs); i += batchSize {
		end := i + batchSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		jobs = append(jobs, processJob{orderIDs: orderIDs[i:end]})
	}

	pool := NewWorkerPool(client, 1, desc.ActionType_ACTION_TYPE_ISSUE)
	pool.Start(2)

	go func() {
		for _, job := range jobs {
			pool.Submit(job)
		}
	}()

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Updating to 3 workers")
		pool.Update(3)

		time.Sleep(2 * time.Second)
		log.Println("Updating to 1 workers")
		pool.Update(1)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		time.Sleep(10 * time.Second)
		sigs <- syscall.SIGINT
	}()

	select {
	case sig := <-sigs:
		log.Printf("Signal %v", sig)
		pool.Shutdown()
	case <-time.After(20 * time.Second):
		log.Println("Timeout")
		pool.Shutdown()
	}
}

func addReadMetadata(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "mode", "read")
}

func addWriteMetadata(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "mode", "write")
}

func acceptOrder(ctx context.Context, client desc.NotifierClient, orderID uint64, userID uint64, expiresAt time.Time, pkg *desc.PackageType, weight float32, price float32) error {
	ctx = addWriteMetadata(ctx)

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

func listOrders(ctx context.Context, client desc.NotifierClient, userID uint64, inPvz bool, lastN *uint32, page uint32, limit uint32) error {
	ctx = addReadMetadata(ctx)

	req := &desc.ListOrdersRequest{
		UserId: userID,
		InPvz:  inPvz,
		LastN:  lastN,
		Pagination: &desc.Pagination{
			Page:        page,
			CountOnPage: limit,
		},
	}

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
	ctx = addWriteMetadata(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "sender", "go-client", "client-version", "1.0")

	//todo: замедление
	time.Sleep(1 * time.Second)

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
	ctx = addWriteMetadata(ctx)
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
	ctx = addReadMetadata(ctx)
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
	ctx = addReadMetadata(ctx)
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

func getOrderHistory(ctx context.Context, client desc.NotifierClient, orderID uint64) error {
	ctx = addReadMetadata(ctx)
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
