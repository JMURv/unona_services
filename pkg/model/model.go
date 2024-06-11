package model

import (
	"github.com/google/uuid"
	"time"
)

type Notification struct {
	ID           uint64    `json:"id" gorm:"primaryKey"`
	Type         string    `json:"type"`
	UserUUID     uuid.UUID `json:"user_uuid"`
	ReceiverUUID uuid.UUID `json:"receiver_uuid"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"createdAt"`
	ForBoth      bool      `json:"for_both"`
}
