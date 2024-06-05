package model

import (
	"time"
)

type Notification struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	Type       string    `json:"type"`
	UserID     uint64    `json:"user_id"`
	ReceiverID uint64    `json:"receiver_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
}
