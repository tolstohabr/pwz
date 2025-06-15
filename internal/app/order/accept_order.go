package order

import (
	"context"

	"PWZ1.0/internal/models"
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

func toInternalPackage(pt desc.PackageType) models.PackageType {
	switch pt {
	case desc.PackageType_PACKAGE_TYPE_BAG:
		return models.PackageBag
	case desc.PackageType_PACKAGE_TYPE_BOX:
		return models.PackageBox
	case desc.PackageType_PACKAGE_TYPE_TAPE:
		return models.PackageFilm
	case desc.PackageType_PACKAGE_TYPE_BAG_TAPE:
		return models.PackageBagFilm
	case desc.PackageType_PACKAGE_TYPE_BOX_TAPE:
		return models.PackageBoxFilm
	default:
		return models.PackageNone
	}
}
