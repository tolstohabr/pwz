package order

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

	"PWZ1.0/pkg/pwz"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func (i *Implementation) GetHistory(ctx context.Context, req *pwz.GetHistoryRequest) (*pwz.OrderHistoryList, error) {
	file, err := os.Open("order_history.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var historyList []*pwz.OrderHistory

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var record struct {
			OrderID   uint64 `json:"order_id"`
			Status    string `json:"status"`
			Timestamp string `json:"created_at"`
		}

		if err := json.Unmarshal([]byte(line), &record); err != nil {
			continue
		}

		createdAt, err := time.Parse(time.RFC3339, record.Timestamp)
		if err != nil {
			continue
		}

		ts := &timestamp.Timestamp{
			Seconds: createdAt.Unix(),
			Nanos:   int32(createdAt.Nanosecond()),
		}

		historyList = append(historyList, &pwz.OrderHistory{
			OrderId:   record.OrderID,
			Status:    pwz.OrderStatus(pwz.OrderStatus_value[record.Status]), // Конвертируем статус
			CreatedAt: ts,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if req.GetPagination() != nil {
		page := req.GetPagination().GetPage()
		perPage := req.GetPagination().GetCountOnPage()
		if perPage > 0 {
			start := (page - 1) * perPage
			end := start + perPage
			if start >= uint32(len(historyList)) {
				historyList = []*pwz.OrderHistory{}
			} else if end > uint32(len(historyList)) {
				historyList = historyList[start:]
			} else {
				historyList = historyList[start:end]
			}
		}
	}

	return &pwz.OrderHistoryList{
		History: historyList,
	}, nil
}
