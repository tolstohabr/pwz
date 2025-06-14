package order

import (
	"context"

	"PWZ1.0/internal/models"
	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) ListOrders(ctx context.Context, req *desc.ListOrdersRequest) (*desc.OrdersList, error) {
	// Извлекаем поля из запроса
	userID := req.GetUserId()
	inPvzOnly := req.GetInPvz()
	lastN := int(req.GetLastN())

	// Пагинация
	page := 0
	limit := 0
	if req.Pagination != nil {
		page = int(req.Pagination.GetPage())
		limit = int(req.Pagination.GetCountOnPage())
	}

	// Получаем список заказов через бизнес-логику
	orders := i.orderService.ListOrders(ctx, userID, inPvzOnly, lastN, page, limit)

	// Преобразуем внутренние заказы в proto-заказы
	var protoOrders []*desc.Order
	for _, o := range orders {
		protoOrder := &desc.Order{
			OrderId:    o.ID,
			UserId:     o.UserID,
			Status:     toProtoStatus(o.Status),
			ExpiresAt:  timestamppb.New(o.ExpiresAt),
			Weight:     o.Weight,
			TotalPrice: o.Price,
		}

		if o.PackageType != models.PackageNone {
			protoOrder.Package = toProtoPackage(o.PackageType)
		}

		protoOrders = append(protoOrders, protoOrder)
	}

	return &desc.OrdersList{
		Orders: protoOrders,
		Total:  int32(len(protoOrders)), // или общее число до пагинации, если нужно
	}, nil
}

func toProtoStatus(s models.OrderStatus) desc.OrderStatus {
	switch s {
	case models.StatusAccepted:
		return desc.OrderStatus_ORDER_STATUS_EXPECTS
	case models.StatusIssued:
		return desc.OrderStatus_ORDER_STATUS_ACCEPTED
	case models.StatusReturned:
		return desc.OrderStatus_ORDER_STATUS_RETURNED
	default:
		return desc.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func toProtoPackage(p models.PackageType) *desc.PackageType {
	var pt desc.PackageType
	switch p {
	case models.PackageBag:
		pt = desc.PackageType_PACKAGE_TYPE_BAG
	case models.PackageBox:
		pt = desc.PackageType_PACKAGE_TYPE_BOX
	case models.PackageFilm:
		pt = desc.PackageType_PACKAGE_TYPE_TAPE
	case models.PackageBagFilm:
		pt = desc.PackageType_PACKAGE_TYPE_BAG_TAPE
	case models.PackageBoxFilm:
		pt = desc.PackageType_PACKAGE_TYPE_BOX_TAPE
	default:
		pt = desc.PackageType_PACKAGE_TYPE_UNSPECIFIED
	}
	return &pt
}
