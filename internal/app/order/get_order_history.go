package order

import (
	"context"

	"PWZ1.0/internal/models/domainErrors"
	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) GetOrderHistory(ctx context.Context, req *desc.OrderHistoryRequest) (*desc.OrderHistoryResponse, error) {
	orderID := req.GetOrderId()

	history, err := i.orderService.GetOrderHistory(ctx, orderID)
	if err != nil {
		if err == domainErrors.ErrOrderNotFound {
			return nil, status.Errorf(codes.NotFound, "Order not found")
		}
		return nil, status.Errorf(codes.Internal, "Internal error: %v", err)
	}

	resp := &desc.OrderHistoryResponse{}
	for _, h := range history {
		resp.History = append(resp.History, &desc.OrderHistory{
			OrderId:   h.OrderID,
			Status:    convertOrderStatus(h.Status),
			CreatedAt: timestamppb.New(h.CreatedAt),
		})
	}
	return resp, nil
}
