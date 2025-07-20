package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OrdersIssued = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_issued_total",
			Help: "number of orders issued",
		},
	)
)

func Init() {
	prometheus.MustRegister(OrdersIssued)
}
