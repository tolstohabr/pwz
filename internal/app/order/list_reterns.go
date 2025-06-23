package order

import (
	"context"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/service"
	"PWZ1.0/pkg/pwz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) ListReturns(ctx context.Context, req *pwz.ListReturnsRequest) (*pwz.ReturnsList, error) {
	serviceReq := service.ListReturnsRequest{
		Pagination: service.Pagination{
			Page:        req.Pagination.GetPage(),
			CountOnPage: req.Pagination.GetCountOnPage(),
		},
	}

	serviceResp := i.orderService.ListReturns(ctx, serviceReq)

	var returns []*pwz.Order
	for _, o := range serviceResp.Returns {
		order := &pwz.Order{
			OrderId:    o.ID,
			UserId:     o.UserID,
			Status:     convertStatusToProto(o.Status),
			ExpiresAt:  timestamppb.New(o.ExpiresAt),
			Weight:     o.Weight,
			TotalPrice: o.Price,
		}

		if o.PackageType != models.PackageUnspecified {
			pkgType := convertPackageToProto(o.PackageType)
			order.Package = &pkgType
		}

		returns = append(returns, order)
	}

	return &pwz.ReturnsList{
		Returns: returns,
	}, nil
}
