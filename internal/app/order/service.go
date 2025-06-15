package order

import (
	"context"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/service"
	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Implementation struct {
	desc.UnimplementedNotifierServer
	orderService service.OrderService
}

func NewHandler(orderService service.OrderService) *Implementation {
	return &Implementation{orderService: orderService}
}

func (i *Implementation) ListOrders(ctx context.Context, req *desc.ListOrdersRequest) (*desc.OrdersList, error) {
	// Преобразование параметров запроса
	var lastId uint32
	if req.LastN != nil {
		lastId = *req.LastN
	}

	var page, limit uint32
	if req.Pagination != nil {
		page = req.Pagination.Page
		limit = req.Pagination.CountOnPage
	} else {
		// Значения по умолчанию для пагинации
		page = 0
		limit = 10
	}

	// Вызов сервиса
	orders, total := i.orderService.ListOrders(ctx, req.UserId, req.InPvz, lastId, page, limit)

	// Преобразование результатов в protobuf формат
	pbOrders := make([]*desc.Order, 0, len(orders))
	for _, o := range orders {
		pbOrder := &desc.Order{
			OrderId:    o.ID,
			UserId:     o.UserID,
			Status:     mapOrderStatusToPb(o.Status),
			ExpiresAt:  timestamppb.New(o.ExpiresAt),
			Weight:     o.Weight,
			TotalPrice: o.Price,
		}

		if o.PackageType != "" && o.PackageType != models.PackageNone {
			pbPackage := mapPackageTypeToPb(o.PackageType)
			pbOrder.Package = &pbPackage
		}

		pbOrders = append(pbOrders, pbOrder)
	}

	return &desc.OrdersList{
		Orders: pbOrders,
		Total:  int32(total),
	}, nil
}

// Вспомогательные функции для преобразования типов
func mapOrderStatusToPb(status models.OrderStatus) desc.OrderStatus {
	switch status {
	case models.StatusAccepted:
		return desc.OrderStatus_ORDER_STATUS_ACCEPTED
	case models.StatusIssued:
		return desc.OrderStatus_ORDER_STATUS_EXPECTS // предположение, что ISSUED соответствует EXPECTS
	case models.StatusReturned:
		return desc.OrderStatus_ORDER_STATUS_RETURNED
	default:
		return desc.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func mapPackageTypeToPb(pkgType models.PackageType) desc.PackageType {
	switch pkgType {
	case models.PackageBag:
		return desc.PackageType_PACKAGE_TYPE_BAG
	case models.PackageBox:
		return desc.PackageType_PACKAGE_TYPE_BOX
	case models.PackageFilm:
		return desc.PackageType_PACKAGE_TYPE_TAPE
	case models.PackageBagFilm:
		return desc.PackageType_PACKAGE_TYPE_BAG_TAPE
	case models.PackageBoxFilm:
		return desc.PackageType_PACKAGE_TYPE_BOX_TAPE
	default:
		return desc.PackageType_PACKAGE_TYPE_UNSPECIFIED
	}
}
