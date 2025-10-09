package models

import "time"

type GlobalFlag struct {
	Key       string `gorm:"primaryKey"`
	Value     string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
