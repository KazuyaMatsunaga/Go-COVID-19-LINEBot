package entity

import (
	"time"
)

type Covid19Info struct {
	ID        int64 `gorm:"primary_key;not null"`
	Cases     int
	Deaths    int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
