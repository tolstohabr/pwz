package order

import (
	"context"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/service"
	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) GetHistory(ctx context.Context, req *desc.GetHistoryRequest) (*desc.OrderHistoryList, error) {
	serviceReq := service.GetHistoryRequest{
		Pagination: service.Pagination{
			Page:        req.GetPagination().GetPage(),
			CountOnPage: req.GetPagination().GetCountOnPage(),
		},
	}

	historyList := i.orderService.GetHistory(serviceReq)

	var pbHistoryList []*desc.OrderHistory
	for _, h := range historyList.History {
		var status desc.OrderStatus
		switch h.Status {
		case models.StatusExpects:
			status = desc.OrderStatus_ORDER_STATUS_EXPECTS
		case models.StatusAccepted:
			status = desc.OrderStatus_ORDER_STATUS_ACCEPTED
		case models.StatusReturned:
			status = desc.OrderStatus_ORDER_STATUS_RETURNED
		case models.StatusDeleted:
			status = desc.OrderStatus_ORDER_STATUS_DELETED
		default:
			status = desc.OrderStatus_ORDER_STATUS_UNSPECIFIED
		}

		pbHistoryList = append(pbHistoryList, &desc.OrderHistory{
			OrderId:   h.OrderID,
			Status:    status,
			CreatedAt: timestamppb.New(h.CreatedAt),
		})
	}

	return &desc.OrderHistoryList{
		History: pbHistoryList,
	}, nil
}
