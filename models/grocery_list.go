package models

import "time"

const (
	EnumDefaultGroceryList = -1
)

type GroceryList struct {
	ID        uint   `gorm:"primaryKey"`
	GuildID   string `gorm:"not null;index"`
	ListLabel string `gorm:"not null;index"`
	FancyName *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (gl *GroceryList) GetName() string {
	if gl == nil {
		return "your grocery list"
	}
	if gl.FancyName != nil {
		return *gl.FancyName
	}
	return gl.ListLabel
}
