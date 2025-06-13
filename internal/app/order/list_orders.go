package order

import (
	"context"
	"strconv"

	"PWZ1.0/internal/models"
	desc "PWZ1.0/pkg/order"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Implementation) ListOrders(ctx context.Context, req *desc.ListOrdersRequest) (*desc.OrdersList, error) {
	// вызов бизнес-логики через s.orderService
	userID := req.UserId
	inPvzOnly := req.InPvz
	var lastCount, page, limit int
	if req.LastN != nil {
		lastCount = int(*req.LastN)
	}
	if req.Pagination != nil {
		page = int(req.Pagination.Page)
		limit = int(req.Pagination.CountOnPage)
	}

	//TODO: костыль ниже чтобы из int перевести в tring
	userIDstr := strconv.FormatUint(userID, 10)
	orders := s.orderService.ListOrders(ctx, userIDstr, inPvzOnly, lastCount, page, limit)

	// Преобразуй domain-модели orders в protobuf OrdersList
	// (примерно, нужно реализовать маппинг)

	var pbOrders []*desc.Order
	for _, o := range orders {
		pbOrders = append(pbOrders, &desc.Order{
			OrderId:    o.ID,
			UserId:     o.UserID,
			Status:     desc.OrderStatus(o.Status),
			Weight:     float32(o.Weight),
			TotalPrice: float32(o.Price),
			// и другие поля, преобразуй по необходимости
		})
	}

	return &desc.OrdersList{
		Orders: pbOrders,
		Total:  int32(len(pbOrders)),
	}, nil
}
