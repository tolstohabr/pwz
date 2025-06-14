package order

import (
	"context"

	desc "PWZ1.0/pkg/pwz"
)

func (i *Implementation) AcceptOrder(ctx context.Context, req *desc.AcceptOrderRequest) (*desc.OrderResponse, error) {
	expiresAt := req.GetExpiresAt().AsTime()

	order, err := i.orderService.AcceptOrder(
		ctx,
		req.GetOrderId(),
		req.GetUserId(),
		req.GetWeight(),
		req.GetPrice(),
		expiresAt,
		toInternalPackage(req.GetPackage()),
	)

	if err != nil {
		return nil, err
	}

	return &desc.OrderResponse{
		Status:  desc.OrderStatus_ORDER_STATUS_ACCEPTED,
		OrderId: order.ID,
	}, nil
}
