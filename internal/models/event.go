package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID   uuid.UUID  `json:"event_id"`
	EventType string     `json:"event_type"`
	Timestamp time.Time  `json:"timestamp"`
	Actor     Actor      `json:"actor"`
	Order     EventOrder `json:"order"`
	Source    string     `json:"source"`
}

type Actor struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type EventOrder struct {
	ID     uint64      `json:"id"`
	UserID uint64      `json:"user_id"`
	Status OrderStatus `json:"status"`
}
