package models

import (
	"time"

	"PWZ1.0/internal/models/domainErrors"
)

type OrderStatus string
type PackageType string

const (
	//статусы
	StatusAccepted OrderStatus = "ACCEPTED" //заказ принят от курьера и лежит на складе
	StatusIssued   OrderStatus = "ISSUED"   //заказ у клиента
	StatusReturned OrderStatus = "RETURNED" //заказ возвращен
	//виды упаковки
	PackageBag     PackageType = "bag"      //пакет
	PackageBox     PackageType = "box"      //коробка
	PackageFilm    PackageType = "film"     //пленка
	PackageBagFilm PackageType = "bag+film" //пакет и пленка
	PackageBoxFilm PackageType = "box+film" //коробка и пленка
	PackageNone    PackageType = "none"     //нету и пленка
)

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	ExpiresAt time.Time   `json:"expires_at"` //время до которого заказ можно выдать
	Status    OrderStatus `json:"status"`
	IssuedAt  *time.Time  `json:"issued_at,omitempty"` //время когда заказ был выдан клиенту
	//новые поля
	PackageType PackageType `json:"package_type"`
	Weight      float64     `json:"weight"`
	Price       float64     `json:"price"`
}

// расчёт всей стоимости
func (o *Order) CalculateTotalPrice() {
	switch o.PackageType {
	case PackageBag:
		o.Price += 5
	case PackageBox:
		o.Price += 20
	case PackageFilm:
		o.Price += 1
	case PackageBagFilm:
		o.Price += 6
	case PackageBoxFilm:
		o.Price += 21
	}
}

// валидация веса
func (o *Order) ValidationWeight() error {
	switch o.PackageType {
	case PackageNone:
		return nil
	case PackageBag:
		if o.Weight >= 10 {
			return domainErrors.ErrWeightTooHeavy
		}
		return nil
	case PackageBox:
		if o.Weight >= 30 {
			return domainErrors.ErrWeightTooHeavy
		}
		return nil
	case PackageFilm:
		return nil
	case PackageBagFilm:
		if o.Weight >= 10 {
			return domainErrors.ErrWeightTooHeavy
		}
		return nil
	case PackageBoxFilm:
		if o.Weight >= 30 {
			return domainErrors.ErrWeightTooHeavy
		}
		return nil
	default:
		return domainErrors.ErrInvalidPackage
	}
}
