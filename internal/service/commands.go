package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	//"PWZ1.0/internal/models"
	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
	"PWZ1.0/internal/storage"
	"PWZ1.0/internal/tools/logger"
)

const (
	expiredTime    = 48 * time.Hour
	DateTimeFormat = "2006-01-02 15:04:05"
)

type OrderService interface {
	AcceptOrder(ctx context.Context, orderID, userID string, weight, price float64, expiresAt time.Time, packageType models.PackageType) (models.Order, error)
	ReturnOrder(orderID string) error
	ProcessOrders(ctx context.Context, userID string, action string, orderIDs []string) []string
	ListOrders(ctx context.Context, userID string, inPvzOnly bool, lastCount, page, limit int) []models.Order
	ListReturns(page, limit int) []models.Order
	ScrollOrders(userID, lastID string, limit int) ([]models.Order, string)
	SaveOrder(order models.Order) error
}

type orderService struct {
	storage storage.Storage
}

func NewOrderService(storage storage.Storage) OrderService {
	return &orderService{storage: storage}
}

func (s *orderService) SaveOrder(order models.Order) error {
	return s.storage.SaveOrder(order)
}

// AcceptOrder добавить заказ в ПВЗ
func (s *orderService) AcceptOrder(ctx context.Context, orderID, userID string, weight, price float64, expiresAt time.Time, package_type models.PackageType) (models.Order, error) {
	newOrder := models.Order{
		ID:          orderID,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		Status:      models.StatusAccepted,
		Weight:      weight,
		Price:       price,
		PackageType: package_type,
	}

	//валидна ли упаковка
	if !IsValidPackage(package_type) {
		return newOrder, domainErrors.ErrInvalidPackage
	}

	//если срок хранения в прошлом
	if expiresAt.Before(time.Now()) {
		return newOrder, domainErrors.ErrValidationFailed
	}

	//если такой заказ уже есть
	_, err := s.storage.GetOrder(orderID)
	if err == nil {
		return newOrder, domainErrors.ErrOrderAlreadyExists
	}

	//валидация веса
	err = newOrder.ValidationWeight()
	if err != nil {
		return newOrder, err
	}

	//расчёт всей стоимости
	newOrder.CalculateTotalPrice()

	appendToHistory(ctx, orderID, models.StatusAccepted)

	return newOrder, s.storage.SaveOrder(newOrder)
}

// IsValidPackage валидна ли упаковка
func IsValidPackage(pkg models.PackageType) bool {
	switch pkg {
	case models.PackageBag, models.PackageBox, models.PackageFilm, models.PackageBagFilm, models.PackageBoxFilm, models.PackageNone:
		return true
	}
	return false
}

// ReturnOrder удалить заказ
func (s *orderService) ReturnOrder(orderID string) error {
	order, err := s.storage.GetOrder(orderID)
	if err != nil {
		return err
	}

	//если заказ у клиента
	if order.Status == models.StatusIssued {
		return domainErrors.ErrOrderAlreadyIssued
	}

	//если заказ в ПВЗ после возврата
	if order.Status == models.StatusReturned {
		return s.storage.DeleteOrder(orderID)
	}

	//если время хранения не истекло
	if time.Now().Before(order.ExpiresAt) {
		return domainErrors.ErrStorageNotExpired
	}

	return s.storage.DeleteOrder(orderID)
}

// ProcessOrders обработать выдачу или возврат заказа
func (s *orderService) ProcessOrders(ctx context.Context, userID string, action string, orderIDs []string) []string {
	var results []string
	for _, id := range orderIDs {
		order, err := s.storage.GetOrder(id)
		if err != nil {
			results = append(results, fmt.Sprintf("ERROR %s: ORDER_NOT_FOUND: заказ не найден", id))
			continue
		}

		if order.UserID != userID {
			results = append(results, fmt.Sprintf("ERROR %s: USER_MISMATCH: несоответствие ID", id))
			continue
		}

		if action == "issue" {
			if time.Now().After(order.ExpiresAt) {
				results = append(results, fmt.Sprintf("ERROR %s: STORAGE_EXPIRED: срок хранения истёк", id))
				continue
			}
			now := time.Now()
			order.Status = models.StatusIssued
			order.IssuedAt = &now
			appendToHistory(ctx, order.ID, models.StatusIssued)
		} else if action == "return" {
			if order.IssuedAt == nil || time.Since(*order.IssuedAt) > expiredTime {
				results = append(results, fmt.Sprintf("ERROR %s: RETURN_TIME_EXPIRED: время на возврат истекло", id))
				continue
			}
			order.Status = models.StatusReturned
			appendToHistory(ctx, order.ID, models.StatusReturned)
		} else {
			results = append(results, fmt.Sprintf("ERROR %s: INVALID_ACTION: непредусмотренное действие", id))
			continue
		}

		_ = s.storage.UpdateOrder(order)

		results = append(results, fmt.Sprintf("PROCESSED: %s", id))
	}
	return results
}

// ListOrders вывести список заказов
func (s *orderService) ListOrders(ctx context.Context, userID string, inPvzOnly bool, lastCount int, page int, limit int) []models.Order {
	if limit <= 0 {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "limit должен быть больше нуля")
		return nil
	}

	allOrders, err := s.storage.ListOrders()
	if err != nil {
		return []models.Order{}
	}

	filtered := make([]models.Order, 0)
	for _, o := range allOrders {
		if o.UserID != userID {
			continue
		}
		if inPvzOnly {
			if o.Status != models.StatusAccepted && o.Status != models.StatusReturned {
				continue
			}
		}
		filtered = append(filtered, o)
	}

	if lastCount > 0 && lastCount < len(filtered) {
		filtered = filtered[len(filtered)-lastCount:]
	}

	//if limit > 0 {
	start := page * limit
	end := start + limit
	if start >= len(filtered) {
		return []models.Order{}
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	filtered = filtered[start:end]

	return filtered
}

// ListReturns вывести список возвратов
func (s *orderService) ListReturns(page int, limit int) []models.Order {
	allOrders, err := s.storage.ListOrders()
	if err != nil {
		return []models.Order{}
	}

	var returned []models.Order
	for _, o := range allOrders {
		if o.Status == models.StatusReturned {
			returned = append(returned, o)
		}
	}

	if limit > 0 {
		start := page * limit
		end := start + limit
		if start >= len(returned) {
			return []models.Order{}
		}
		if end > len(returned) {
			end = len(returned)
		}
		return returned[start:end]
	}

	return returned
}

// appendToHistory для добавления записи об изменении статуса в json-ку
func appendToHistory(ctx context.Context, orderID string, status models.OrderStatus) {
	record := map[string]string{
		"order_id":  orderID,
		"status":    string(status),
		"timestamp": time.Now().Format(DateTimeFormat),
	}

	file, err := os.OpenFile("order_history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrOpenFiled, "ошибка открытия")
		return
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrReadFiled, "ошибка записи")
		return
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrReadFiled, "ошибка записи")
		return
	}
}

// ScrollOrders прокрутка
func (s *orderService) ScrollOrders(userID string, lastID string, limit int) ([]models.Order, string) {
	allOrders, err := s.storage.ListOrders()
	if err != nil {
		return []models.Order{}, ""
	}

	userOrders := make([]models.Order, 0)
	for _, o := range allOrders {
		if o.UserID == userID {
			userOrders = append(userOrders, o)
		}
	}

	sort.Slice(userOrders, func(i, j int) bool {
		return userOrders[i].ID < userOrders[j].ID
	})

	// найти индекс последнего заказа с lastID
	startIdx := 0
	if lastID != "0" && lastID != "" {
		for i, o := range userOrders {
			if o.ID == lastID {
				startIdx = i + 1
				break
			}
		}
	}

	// взять следующую пачку
	endIdx := startIdx + limit
	if endIdx > len(userOrders) {
		endIdx = len(userOrders)
	}

	result := userOrders[startIdx:endIdx]

	nextLastID := ""
	if len(result) > 0 {
		nextLastID = result[len(result)-1].ID
	}

	return result, nextLastID
}
