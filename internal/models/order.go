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

func (o *Order) CalculateTotalPrice() error {
	switch o.PackageType {
	case PackageNone:
		return nil
	case PackageBag:
		if o.Weight >= 10 {
			//o.PackageType = PackageNone
			return domainErrors.ErrWeightTooHeavy
		}
		o.Price += 5
		return nil
	case PackageBox:
		if o.Weight >= 30 {
			//o.PackageType = PackageNone
			return domainErrors.ErrWeightTooHeavy
		}
		o.Price += 20
		return nil
	case PackageFilm:
		o.Price += 1
		return nil
	case PackageBagFilm:
		if o.Weight >= 10 {
			//o.PackageType = PackageNone
			return domainErrors.ErrWeightTooHeavy
		}
		o.Price += 6
		return nil
	case PackageBoxFilm:
		if o.Weight >= 30 {
			//o.PackageType = PackageNone
			return domainErrors.ErrWeightTooHeavy
		}
		o.Price += 21
		return nil
	default:
		return domainErrors.ErrInvalidPackage
	}
}
