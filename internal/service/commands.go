package service

import (
	"context"
	"log"
	"sort"
	"time"

	"PWZ1.0/internal/models"
	"PWZ1.0/internal/models/domainErrors"
	"PWZ1.0/internal/storage"
	"PWZ1.0/internal/tools/logger"
	"github.com/jackc/pgx/v5"
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
	ReturnOrder(ctx context.Context, orderID uint64) (*OrderResponse, error)
	ProcessOrders(ctx context.Context, userID uint64, action models.ActionType, orderIDs []uint64) ProcessResult
	ListOrders(ctx context.Context, userID uint64, inPvzOnly bool, lastId, page, limit uint32) ([]models.Order, uint32)
	ListReturns(ctx context.Context, req ListReturnsRequest) ReturnsList
	ScrollOrders(ctx context.Context, userID, lastID uint64, limit int) ([]models.Order, uint64)
	GetHistory(ctx context.Context, page uint32, count uint32) ([]models.OrderHistory, error)
	GetOrderHistory(ctx context.Context, orderID uint64) ([]models.OrderHistory, error)
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

func (s *orderService) AcceptOrder(ctx context.Context, orderID, userID uint64, weight, price float32, expiresAt time.Time, packageType models.PackageType) (models.Order, error) {
	log.Printf("AcceptOrder called: orderID=%d, userID=%d", orderID, userID)

	newOrder := models.Order{
		ID:          orderID,
		UserID:      userID,
		ExpiresAt:   expiresAt,
		Status:      models.StatusExpects,
		Weight:      weight,
		Price:       price,
		PackageType: packageType,
	}

	if !IsValidPackage(packageType) {
		logger.LogErrorWithCode(ctx, domainErrors.ErrInvalidPackage, "Invalid package type")
		return newOrder, domainErrors.ErrInvalidPackage
	}

	if expiresAt.Before(time.Now()) {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "Expiration date in past")
		return newOrder, domainErrors.ErrValidationFailed
	}

	_, err := s.storage.GetOrder(ctx, orderID)
	if err == nil {
		logger.LogErrorWithCode(ctx, domainErrors.ErrOrderAlreadyExists, "Order already exists")
		return newOrder, domainErrors.ErrOrderAlreadyExists
	}

	if err := newOrder.ValidationWeight(); err != nil {
		logger.LogErrorWithCode(ctx, err, "Weight validation failed")
		return newOrder, err
	}

	newOrder.CalculateTotalPrice()

	err = s.storage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		//было
		return s.storage.SaveOrderTx(ctx, tx, newOrder)
		//TODO: надо сделать так
		//s.storage.SaveOrderTx(ctx, tx, newOrder)
		//return s.storage.SaveEventTx(ctx, tx, newOrder)
	})
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to save order")
		return newOrder, err
	}

	log.Printf("Order accepted successfully: orderID=%d", orderID)
	return newOrder, nil
}

func IsValidPackage(pkg models.PackageType) bool {
	switch pkg {
	case models.PackageBag, models.PackageBox, models.PackageTape, models.PackageBagTape, models.PackageBoxTape, models.PackageUnspecified:
		return true
	}
	return false
}

func (s *orderService) ReturnOrder(ctx context.Context, orderID uint64) (*OrderResponse, error) {
	log.Printf("ReturnOrder called: orderID=%d", orderID)

	order, err := s.storage.GetOrder(ctx, orderID)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to get order for return")
		return nil, err
	}

	if order.Status == models.StatusAccepted {
		logger.LogErrorWithCode(ctx, domainErrors.ErrOrderAlreadyIssued, "Order already issued")
		return nil, domainErrors.ErrOrderAlreadyIssued
	}

	if order.Status == models.StatusReturned {
		err := s.storage.DeleteOrder(ctx, orderID)
		if err != nil {
			logger.LogErrorWithCode(ctx, err, "Failed to delete returned order")
			return nil, err
		}
		log.Printf("Returned order deleted: orderID=%d", orderID)
		return &OrderResponse{
			OrderID: orderID,
			Status:  models.StatusDeleted,
		}, nil
	}

	if time.Now().Before(order.ExpiresAt) {
		logger.LogErrorWithCode(ctx, domainErrors.ErrStorageNotExpired, "Storage not expired yet")
		return nil, domainErrors.ErrStorageNotExpired
	}

	err = s.storage.DeleteOrder(ctx, orderID)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to delete expired order")
		return nil, err
	}

	log.Printf("Expired order deleted: orderID=%d", orderID)
	return &OrderResponse{
		OrderID: orderID,
		Status:  models.StatusDeleted,
	}, nil
}

func (s *orderService) ProcessOrders(ctx context.Context, userID uint64, actionType models.ActionType, orderIDs []uint64) ProcessResult {
	log.Printf("ProcessOrders called: userID=%d, action=%v, orderIDs=%v", userID, actionType, orderIDs)

	result := ProcessResult{
		Processed: make([]uint64, 0),
		Errors:    make([]uint64, 0),
	}

	err := s.storage.WithTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		for _, id := range orderIDs {
			order, err := s.storage.GetOrder(ctx, id)
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

			case models.ActionTypeReturn:
				if order.Status != models.StatusAccepted {
					result.Errors = append(result.Errors, id)
					continue
				}
				order.Status = models.StatusReturned

			default:
				result.Errors = append(result.Errors, id)
				continue
			}

			err = s.storage.UpdateOrderTx(ctx, tx, order)
			if err != nil {
				result.Errors = append(result.Errors, id)
				continue
			}
			result.Processed = append(result.Processed, id)
		}
		return nil
	})

	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Transaction error in ProcessOrders")
	}

	log.Printf("ProcessOrders result: processed=%d, errors=%d", len(result.Processed), len(result.Errors))
	return result
}

func (s *orderService) ListOrders(ctx context.Context, userID uint64, inPvzOnly bool, lastId uint32, page uint32, limit uint32) ([]models.Order, uint32) {
	log.Printf("ListOrders called: userID=%d, inPvzOnly=%v, lastId=%d, page=%d, limit=%d",
		userID, inPvzOnly, lastId, page, limit)

	if limit == 0 {
		logger.LogErrorWithCode(ctx, domainErrors.ErrValidationFailed, "Limit must be greater than zero")
		return nil, 0
	}

	allOrders, err := s.storage.ListOrders(ctx)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to list orders")
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

	log.Printf("ListOrders result: count=%d", len(filtered))
	return filtered, total
}

func (s *orderService) ListReturns(ctx context.Context, req ListReturnsRequest) ReturnsList {
	log.Printf("ListReturns called: page=%d, count=%d", req.Pagination.Page, req.Pagination.CountOnPage)

	allOrders, err := s.storage.ListOrders(ctx)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to list returns")
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

	log.Printf("ListReturns result: count=%d", len(returned))
	return ReturnsList{Returns: returned}
}

func (s *orderService) ScrollOrders(ctx context.Context, userID uint64, lastID uint64, limit int) ([]models.Order, uint64) {
	log.Printf("ScrollOrders called: userID=%d, lastID=%d, limit=%d", userID, lastID, limit)

	allOrders, err := s.storage.ListOrders(ctx)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to list orders for scrolling")
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

	log.Printf("ScrollOrders result: count=%d, nextLastID=%d", len(result), nextLastID)
	return result, nextLastID
}

func (s *orderService) GetHistory(ctx context.Context, page, count uint32) ([]models.OrderHistory, error) {
	log.Printf("GetHistory called: page=%d, count=%d", page, count)
	return s.storage.GetHistory(ctx, page, count)
}

func (s *orderService) GetOrderHistory(ctx context.Context, orderID uint64) ([]models.OrderHistory, error) {
	log.Printf("GetOrderHistory called: orderID=%d", orderID)

	history, err := s.storage.GetHistory(ctx, 0, 0)
	if err != nil {
		logger.LogErrorWithCode(ctx, err, "Failed to get order history")
		return nil, err
	}

	var filtered []models.OrderHistory
	for _, h := range history {
		if h.OrderID == orderID {
			filtered = append(filtered, h)
		}
	}

	if len(filtered) == 0 {
		logger.LogErrorWithCode(ctx, domainErrors.ErrOrderNotFound, "Order history not found")
		return nil, domainErrors.ErrOrderNotFound
	}

	log.Printf("GetOrderHistory result: count=%d", len(filtered))
	return filtered, nil
}
