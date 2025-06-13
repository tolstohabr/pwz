package order

import (
	"PWZ1.0/internal/service"
	desc "PWZ1.0/pkg/order"
)

type Implementation struct {
	desc.UnimplementedNotifierServer
	orderService service.OrderService
}

func NewHandler(orderService service.OrderService) *Implementation {
	return &Implementation{orderService: orderService}
}
