package order

import (
	"context"

	desc "PWZ1.0/pkg/pwz"
)

func (i *Implementation) ReturnOrder(ctx context.Context, req *desc.OrderIdRequest) (*desc.OrderResponse, error) {
	orderID := req.GetOrderId()

	resp, err := i.orderService.ReturnOrder(orderID)
	if err != nil {
		return nil, err
	}

	return &desc.OrderResponse{
		Status:  convertStatusToProto(resp.Status),
		OrderId: resp.OrderID,
	}, nil
}
