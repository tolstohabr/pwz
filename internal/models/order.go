package models

import (
	"time"

	"PWZ1.0/internal/models/domainErrors"
)

type OrderStatus string
type PackageType string
type ActionType int

const (
	//статусы
	//старые
	/*StatusExpects OrderStatus = "ACCEPTED" //заказ принят от курьера и лежит на складе
	StatusAccepted   OrderStatus = "ISSUED"   //заказ у клиента
	StatusReturned OrderStatus = "RETURNED" //заказ возвращен*/
	StatusUnspecified OrderStatus = "UNSPECIFIED" // не указан
	StatusExpects     OrderStatus = "EXPECTS"     // получен от курьера, ожидает выдачи клиенту
	StatusAccepted    OrderStatus = "ACCEPTED"    // выдан клиенту
	StatusReturned    OrderStatus = "RETURNED"    // возвращен клиентом в ПВЗ
	StatusDeleted     OrderStatus = "DELETED"     // возвращен курьеру из ПВЗ(удален)

	//виды упаковки
	PackageBag         PackageType = "bag"         // пакет
	PackageBox         PackageType = "box"         // коробка
	PackageTape        PackageType = "tape"        // пленка
	PackageBagTape     PackageType = "bag+tape"    // пакет и пленка
	PackageBoxTape     PackageType = "box+tape"    // коробка и пленка
	PackageUnspecified PackageType = "unspecified" // нету
	//виды действий
	ActionTypeUnspecified ActionType = iota
	ActionTypeIssue
	ActionTypeReturn
)

func (a ActionType) String() string {
	switch a {
	case ActionTypeIssue:
		return "issue"
	case ActionTypeReturn:
		return "return"
	default:
		return "unspecified"
	}
}

// ParseActionType парсит строку в ActionType
func ParseActionType(s string) ActionType {
	switch s {
	case "issue":
		return ActionTypeIssue
	case "return":
		return ActionTypeReturn
	default:
		return ActionTypeUnspecified
	}
}

type Order struct {
	ID          uint64      `json:"id"`
	UserID      uint64      `json:"user_id"`
	ExpiresAt   time.Time   `json:"expires_at"` //время до которого заказ можно выдать
	Status      OrderStatus `json:"status"`
	PackageType PackageType `json:"package_type"`
	Weight      float32     `json:"weight"`
	Price       float32     `json:"price"`
}

// расчёт всей стоимости
func (o *Order) CalculateTotalPrice() {
	switch o.PackageType {
	case PackageBag:
		o.Price += 5
	case PackageBox:
		o.Price += 20
	case PackageTape:
		o.Price += 1
	case PackageBagTape:
		o.Price += 6
	case PackageBoxTape:
		o.Price += 21
	case PackageUnspecified:
		o.Price += 0
	}
}

// валидация веса
func (o *Order) ValidationWeight() error {
	switch o.PackageType {
	case PackageUnspecified:
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
	case PackageTape:
		return nil
	case PackageBagTape:
		if o.Weight >= 10 {
			return domainErrors.ErrWeightTooHeavy
		}
		return nil
	case PackageBoxTape:
		if o.Weight >= 30 {
			return domainErrors.ErrWeightTooHeavy
		}
		return nil
	default:
		return domainErrors.ErrInvalidPackage
	}
}
