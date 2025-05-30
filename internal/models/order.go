package models

import "time"

type OrderStatus string

const (
	StatusAccepted OrderStatus = "ACCEPTED" //заказ принят от курьера и лежит на складе
	StatusIssued   OrderStatus = "ISSUED"   //заказ у клиента
	StatusReturned OrderStatus = "RETURNED" //заказ возвращен
)

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	ExpiresAt time.Time   `json:"expires_at"` //время до которого заказ можно выдать
	Status    OrderStatus `json:"status"`
	IssuedAt  *time.Time  `json:"issued_at,omitempty"` //время когда заказ был выдан клиенту
}
