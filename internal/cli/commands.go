package cli

import (
	"PWZ1.0/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"PWZ1.0/internal/storage"
)

// добавить заказ в ПВЗ
func AcceptOrder(storage storage.Storage, orderID, userID string, expiresAt time.Time) error {
	//если срок хранения в прошлом
	if expiresAt.Before(time.Now()) {
		return errors.New("ERROR: VALIDATION_FAILED: срок хранения в прошлом")
	}

	//если такой заказ уже есть
	_, err := storage.GetOrder(orderID)
	if err == nil {
		return errors.New("ERROR: ORDER_ALREADY_EXISTS: заказ уже есть")
	}

	newOrder := models.Order{
		ID:        orderID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Status:    models.StatusAccepted,
	}

	appendToHistory(orderID, models.StatusAccepted)

	return storage.SaveOrder(newOrder)
}

// удалить заказ
func ReturnOrder(storage storage.Storage, orderID string) error {
	order, err := storage.GetOrder(orderID)
	if err != nil {
		return err
	}

	//если заказ у клиента
	if order.Status == models.StatusIssued {
		return errors.New("ERROR: ORDER_ALREADY_ISSUED: заказ у клиента")
	}

	//если заказ в ПВЗ после возврата
	if order.Status == models.StatusReturned {
		return storage.DeleteOrder(orderID)
	}

	//если время хранения не истекло
	if time.Now().Before(order.ExpiresAt) {
		return errors.New("ERROR: STORAGE_NOT_EXPIRED: время хранения не истекло")
	}

	return storage.DeleteOrder(orderID)
}

// обработать выдачу или возврат заказа
func ProcessOrders(storage storage.Storage, userID string, action string, orderIDs []string) []string {
	var results []string
	for _, id := range orderIDs {
		order, err := storage.GetOrder(id)
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
			appendToHistory(order.ID, models.StatusIssued)
		} else if action == "return" {
			if order.IssuedAt == nil || time.Since(*order.IssuedAt) > 48*time.Hour {
				results = append(results, fmt.Sprintf("ERROR %s: RETURN_TIME_EXPIRED: время на возврат истекло", id))
				continue
			}
			order.Status = models.StatusReturned
			appendToHistory(order.ID, models.StatusReturned)
		} else {
			results = append(results, fmt.Sprintf("ERROR %s: INVALID_ACTION: непредусмотренное действие", id))
			continue
		}

		_ = storage.DeleteOrder(order.ID)
		_ = storage.SaveOrder(order)

		results = append(results, fmt.Sprintf("PROCESSED: %s", id))
	}
	return results
}

// вывести список заказов
func ListOrders(storage storage.Storage, userID string, inPvzOnly bool, lastCount int, page int, limit int) []models.Order {
	allOrders, err := storage.ListOrders()
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

	if limit > 0 {
		start := page * limit
		end := start + limit
		if start >= len(filtered) {
			return []models.Order{}
		}
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	return filtered
}

// вывести список возвратов
func ListReturns(storage storage.Storage, page int, limit int) []models.Order {
	allOrders, err := storage.ListOrders()
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

// для добавления записи об изменении статуса в json-ку
func appendToHistory(orderID string, status models.OrderStatus) {
	record := map[string]string{
		"order_id":  orderID,
		"status":    string(status),
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	file, err := os.OpenFile("order_history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("ERROR: INTERNAL_ERROR:ошибка открытия", err)
		return
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		fmt.Println("ERROR: INTERNAL_ERROR: ошибка с записью:", err)
		return
	}

	file.Write(append(data, '\n'))
}

// прокрутка
func ScrollOrders(storage storage.Storage, userID string, lastID string, limit int) ([]models.Order, string) {
	allOrders, err := storage.ListOrders()
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
