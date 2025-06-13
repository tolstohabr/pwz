package order

import (
	"PWZ1.0/internal/service"
	desc "PWZ1.0/pkg/order" //деском обзывают то что из pkg подтягивают
)

type Implementation struct {
	desc.UnimplementedNotifierServer
	orderService service.OrderService
}

func New() *Implementation {
	return &Implementation{}
}
