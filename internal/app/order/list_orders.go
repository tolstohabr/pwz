package order

import (
	"context"

	"PWZ1.0/internal/models"
	desc "PWZ1.0/pkg/pwz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) ListOrders(ctx context.Context, req *desc.ListOrdersRequest) (*desc.OrdersList, error) {
	var lastId uint32
	if req.LastN != nil {
		lastId = *req.LastN
	}

	var page, limit uint32
	if req.Pagination != nil {
		page = req.Pagination.Page
		limit = req.Pagination.CountOnPage
	} else {
		page = 0
		limit = 10
	}

	orders, total := i.orderService.ListOrders(ctx, req.UserId, req.InPvz, lastId, page, limit)

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

		if o.PackageType != "" && o.PackageType != models.PackageUnspecified {
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

func mapOrderStatusToPb(status models.OrderStatus) desc.OrderStatus {
	switch status {
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

func mapPackageTypeToPb(pkgType models.PackageType) desc.PackageType {
	switch pkgType {
	case models.PackageBag:
		return desc.PackageType_PACKAGE_TYPE_BAG
	case models.PackageBox:
		return desc.PackageType_PACKAGE_TYPE_BOX
	case models.PackageTape:
		return desc.PackageType_PACKAGE_TYPE_TAPE
	case models.PackageBagTape:
		return desc.PackageType_PACKAGE_TYPE_BAG_TAPE
	case models.PackageBoxTape:
		return desc.PackageType_PACKAGE_TYPE_BOX_TAPE
	default:
		return desc.PackageType_PACKAGE_TYPE_UNSPECIFIED
	}
}
