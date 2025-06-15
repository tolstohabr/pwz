package order

import (
	"context"

	"PWZ1.0/internal/models"
	desc "PWZ1.0/pkg/pwz"
)

func (i *Implementation) ProcessOrders(ctx context.Context, req *desc.ProcessOrdersRequest) (*desc.ProcessResult, error) {
	userID := req.GetUserId()
	actionType := convertActionTypeFromProto(req.GetAction())
	orderIDs := req.GetOrderIds()

	result := i.orderService.ProcessOrders(ctx, userID, actionType, orderIDs)

	return &desc.ProcessResult{
		Processed: result.Processed,
		Errors:    result.Errors,
	}, nil
}

func convertActionTypeFromProto(actionType desc.ActionType) models.ActionType {
	switch actionType {
	case desc.ActionType_ACTION_TYPE_ISSUE:
		return models.ActionTypeIssue
	case desc.ActionType_ACTION_TYPE_RETURN:
		return models.ActionTypeReturn
	default:
		return models.ActionTypeUnspecified
	}
}

func convertStatusToProto(status models.OrderStatus) desc.OrderStatus {
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

func convertPackageToProto(pkg models.PackageType) desc.PackageType {
	switch pkg {
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
