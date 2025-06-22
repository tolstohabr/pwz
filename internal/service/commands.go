package service

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"time"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
	"PWZ1.0/internal/storage"
	"PWZ1.0/internal/tools/logger"
)

const (
	ExpiredTime    = 48 * time.Hour
	DateTimeFormat = "2006-01-02 15:04:05"
)

type ListReturnsRequest struct {
	Pagination Pagination
}

type Pagination struct {
	Page        uint32
	CountOnPage uint32
}

type ReturnsList struct {
	Returns []models.Order
}

type OrderService interface {
	AcceptOrder(ctx context.Context, orderID, userID uint64, weight, price float32, expiresAt time.Time, packageType models.PackageType) (models.Order, error)
	ReturnOrder(orderID uint64) (*OrderResponse, error)
	ProcessOrders(ctx context.Context, userID uint64, action models.ActionType, orderIDs []uint64) ProcessResult
	ListOrders(ctx context.Context, userID uint64, inPvzOnly bool, lastId, page, limit uint32) ([]models.Order, uint32)
	ListReturns(req ListReturnsRequest) ReturnsList
	ScrollOrders(userID, lastID uint64, limit int) ([]models.Order, uint64)
	SaveOrder(order models.Order) error
	GetHistory(ctx context.Context, page uint32, count uint32) ([]models.OrderHistory, error)
}

type ProcessResult struct {
	Processed []uint64
	Errors    []uint64
}

type orderService struct {
	storage storage.Storage
}

type OrderResponse struct {
	OrderID uint64
	Status  models.OrderStatus
}

func NewOrderService(storage storage.Storage) OrderService {
	return &orderService{storage: storage}
}

func (s *orderService) SaveOrder(order models.Order) error {
	return s.storage.SaveOrder(order)
}

func (s *orderService) AcceptOrder(ctx context.Context, orderID, userID uint64, weight, price float32, expiresAt time.Time, package_type models.PackageType) (models.Order, error) {
	newOrder := models.Order{
		ID:          orderID,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		Status:      models.StatusExpects,
		Weight:      weight,
		Price:       price,
		PackageType: package_type,
	}

	if !IsValidPackage(package_type) {
		return newOrder, domainErrors.ErrInvalidPackage
	}

	if expiresAt.Before(time.Now()) {
		return newOrder, domainErrors.ErrValidationFailed
	}

	_, err := s.storage.GetOrder(orderID)
	if err == nil {
		return newOrder, domainErrors.ErrOrderAlreadyExists
	}

	err = newOrder.ValidationWeight()
	if err != nil {
		return newOrder, err
	}

	newOrder.CalculateTotalPrice()

	appendToHistory(ctx, orderID, models.StatusExpects)

	return newOrder, s.storage.SaveOrder(newOrder)
}

func IsValidPackage(pkg models.PackageType) bool {
	switch pkg {
	case models.PackageBag, models.PackageBox, models.PackageTape, models.PackageBagTape, models.PackageBoxTape, models.PackageUnspecified:
		return true
	}
	return false
}

func (s *orderService) ReturnOrder(orderID uint64) (*OrderResponse, error) {
	order, err := s.storage.GetOrder(orderID)
	if err != nil {
		return nil, err
	}

	if order.Status == models.StatusAccepted {
		return nil, domainErrors.ErrOrderAlreadyIssued
	}

	if order.Status == models.StatusReturned {
		err := s.storage.DeleteOrder(orderID)
		if err != nil {
			return nil, err
		}
		return &OrderResponse{
			OrderID: orderID,
			Status:  models.StatusDeleted,
		}, nil
	}

	if time.Now().Before(order.ExpiresAt) {
		return nil, domainErrors.ErrStorageNotExpired
	}

	err = s.storage.DeleteOrder(orderID)
	if err != nil {
		return nil, err
	}
	return &OrderResponse{
		OrderID: orderID,
		Status:  models.StatusDeleted,
	}, nil
}

func (s *orderService) ProcessOrders(ctx context.Context, userID uint64, actionType models.ActionType, orderIDs []uint64) ProcessResult {
	result := ProcessResult{
		Processed: make([]uint64, 0),
		Errors:    make([]uint64, 0),
	}

	for _, id := range orderIDs {
		order, err := s.storage.GetOrder(id)
		if err != nil || order.UserID != userID {
			result.Errors = append(result.Errors, id)
			continue
		}

		if time.Now().After(order.ExpiresAt) {
			result.Errors = append(result.Errors, id)
			continue
		}

		switch actionType {
		case models.ActionTypeIssue:
			if order.Status != models.StatusExpects {
				result.Errors = append(result.Errors, id)
				continue
			}
			order.Status = models.StatusAccepted
			order.ExpiresAt = time.Now().Add(ExpiredTime)
			appendToHistory(ctx, order.ID, models.StatusAccepted)

		case models.ActionTypeReturn:
			if order.Status != models.StatusAccepted {
				result.Errors = append(result.Errors, id)
				continue
			}
			order.Status = models.StatusReturned
			appendToHistory(ctx, order.ID, models.StatusReturned)

		default:
			result.Errors = append(result.Errors, id)
			continue
		}

		_ = s.storage.UpdateOrder(order)
		result.Processed = append(result.Processed, id)
	}

	return result
}

func (s *orderService) ListOrders(ctx context.Context, userID uint64, inPvzOnly bool, lastId uint32, page uint32, limit uint32) ([]models.Order, uint32) {
	if limit == 0 {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "limit должен быть больше нуля")
		return nil, 0
	}

	allOrders, err := s.storage.ListOrders()
	if err != nil {
		return []models.Order{}, 0
	}

	filtered := make([]models.Order, 0)
	for _, o := range allOrders {
		if o.UserID != userID {
			continue
		}
		if inPvzOnly {
			if o.Status != models.StatusExpects && o.Status != models.StatusReturned {
				continue
			}
		}
		filtered = append(filtered, o)
	}

	total := uint32(len(filtered))

	if lastId > 0 {
		if lastId > total {
			lastId = total
		}
		filtered = filtered[total-lastId:]
	}

	start := page * limit
	end := start + limit
	if start >= uint32(len(filtered)) {
		return []models.Order{}, total
	}
	if end > uint32(len(filtered)) {
		end = uint32(len(filtered))
	}
	filtered = filtered[start:end]

	return filtered, total
}

func (s *orderService) ListReturns(req ListReturnsRequest) ReturnsList {
	allOrders, err := s.storage.ListOrders()
	if err != nil {
		return ReturnsList{}
	}

	var returned []models.Order
	for _, o := range allOrders {
		if o.Status == models.StatusReturned {
			returned = append(returned, o)
		}
	}

	page := int(req.Pagination.Page)
	limit := int(req.Pagination.CountOnPage)

	if limit > 0 {
		start := page * limit
		end := start + limit
		if start >= len(returned) {
			return ReturnsList{Returns: []models.Order{}}
		}
		if end > len(returned) {
			end = len(returned)
		}
		return ReturnsList{Returns: returned[start:end]}
	}

	return ReturnsList{Returns: returned}
}

func (s *orderService) ScrollOrders(userID uint64, lastID uint64, limit int) ([]models.Order, uint64) {
	allOrders, err := s.storage.ListOrders()
	if err != nil {
		return []models.Order{}, 0
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

	startIdx := 0
	if lastID != 0 {
		for i, o := range userOrders {
			if o.ID == lastID {
				startIdx = i + 1
				break
			}
		}
	}

	endIdx := startIdx + limit
	if endIdx > len(userOrders) {
		endIdx = len(userOrders)
	}

	result := userOrders[startIdx:endIdx]

	var nextLastID uint64
	if len(result) > 0 {
		nextLastID = result[len(result)-1].ID
	}

	return result, nextLastID
}

func appendToHistory(ctx context.Context, orderID uint64, status models.OrderStatus) {
	record := struct {
		OrderID   uint64 `json:"order_id"`
		Status    string `json:"status"`
		Timestamp string `json:"created_at"`
	}{
		OrderID:   orderID,
		Status:    string(status),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	file, err := os.OpenFile("order_history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrOpenFiled, "ошибка открытия")
		return
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrReadFiled, "ошибка маршала")
		return
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrReadFiled, "ошибка записи")
		return
	}
}

type GetHistoryRequest struct {
	Pagination Pagination
}

type OrderHistoryList struct {
	History []OrderHistory
}

type OrderHistory struct {
	OrderID   uint64
	Status    models.OrderStatus
	CreatedAt time.Time
}

func (s *orderService) GetHistory(ctx context.Context, page, count uint32) ([]models.OrderHistory, error) {
	return s.storage.GetHistory(ctx, page, count)
}
