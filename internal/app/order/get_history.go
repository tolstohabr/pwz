package order

import (
	"context"

	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/protobuf/types/known/timestamppb"

	"PWZ1.0/internal/models"
)

func convertOrderStatus(s models.OrderStatus) desc.OrderStatus {
	switch s {
	case models.StatusExpects:
		return desc.OrderStatus_ORDER_STATUS_EXPECTS
	case models.StatusAccepted:
		return desc.OrderStatus_ORDER_STATUS_ACCEPTED
	case models.StatusReturned:
		return desc.OrderStatus_ORDER_STATUS_RETURNED
	case models.StatusDeleted:
		return desc.OrderStatus_ORDER_STATUS_DELETED
	default:
		return desc.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func (i *Implementation) GetHistory(ctx context.Context, req *desc.GetHistoryRequest) (*desc.OrderHistoryList, error) {
	page := uint32(0)
	count := uint32(0)
	if req.GetPagination() != nil {
		page = req.GetPagination().GetPage()
		count = req.GetPagination().GetCountOnPage()
	}

	history, err := i.orderService.GetHistory(ctx, page, count)
	if err != nil {
		return nil, err
	}

	resp := &desc.OrderHistoryList{}
	for _, hItem := range history {
		resp.History = append(resp.History, &desc.OrderHistory{
			OrderId:   hItem.OrderID,
			Status:    convertOrderStatus(hItem.Status),
			CreatedAt: timestamppb.New(hItem.CreatedAt),
		})
	}

	return resp, nil
}
