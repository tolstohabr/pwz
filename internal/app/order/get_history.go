package order

import (
	"context"

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
		pbHistoryList = append(pbHistoryList, &desc.OrderHistory{
			OrderId:   h.OrderID,
			Status:    desc.OrderStatus(desc.OrderStatus_value[string(h.Status)]),
			CreatedAt: timestamppb.New(h.CreatedAt),
		})
	}

	return &desc.OrderHistoryList{
		History: pbHistoryList,
	}, nil
}
