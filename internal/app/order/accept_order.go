package order

import (
	"context"

	desc "PWZ1.0/pkg/order"
)

func (i *Implementation) AcceptOrder(ctx context.Context, req *desc.AcceptOrderRequest) (*desc.OrderResponse, error) {
	// конвертация Timestamp в time.Time
	expiresAt := req.GetExpiresAt().AsTime()

	order, err := i.orderService.AcceptOrder(
		ctx,
		req.GetOrderId(),
		req.GetUserId(),
		float64(req.GetWeight()),
		float64(req.GetPrice()),
		expiresAt,
		toInternalPackage(req.GetPackage()), // преобразуем proto enum -> internal
	)

	if err != nil {
		return nil, err
	}

	return &desc.OrderResponse{
		Status:  desc.OrderStatus_ORDER_STATUS_ACCEPTED,
		OrderId: order.ID,
	}, nil
}
