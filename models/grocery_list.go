package models

import (
	"fmt"
	"time"
)

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

func (gl *GroceryList) GetTitle() string {
	if gl == nil {
		return gl.GetName()
	}
	if gl.FancyName != nil {
		return fmt.Sprintf("%s (%s)", *gl.FancyName, gl.ListLabel)
	} else {
		return fmt.Sprintf("%s", gl.ListLabel)
	}
}

func (gl *GroceryList) GetID() *uint {
	if gl != nil {
		return &gl.ID
	}
	return nil
}

func (gl *GroceryList) GetLabelSuffix() string {
	if gl == nil {
		return ""
	}
	return ":" + gl.ListLabel
}
