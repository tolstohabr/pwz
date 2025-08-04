package order

import (
	"context"

	desc "PWZ1.0/pkg/pwz"
)

func (i *Implementation) ImportOrders(ctx context.Context, req *desc.ImportOrdersRequest) (*desc.ImportResult, error) {
	var (
		importedCount int32
		errorIDs      []uint64
	)

	for _, orderReq := range req.Orders {
		expiresAt := orderReq.GetExpiresAt().AsTime()

		_, err := i.orderService.AcceptOrder(
			ctx,
			orderReq.GetOrderId(),
			orderReq.GetUserId(),
			orderReq.GetWeight(),
			orderReq.GetPrice(),
			expiresAt,
			toInternalPackage(orderReq.GetPackage()),
		)

		if err != nil {
			errorIDs = append(errorIDs, orderReq.GetOrderId())
			continue
		}

		importedCount++
	}

	return &desc.ImportResult{
		Imported: importedCount,
		Errors:   errorIDs,
	}, nil
}
