package models

import (
	"time"
)

// Model contains common fields across all tables.
// It is exactly like gorm.Model with the exception of JSON field tags.
type Model struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
