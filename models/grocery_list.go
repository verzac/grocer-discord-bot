package models

import (
	"fmt"
	"time"
)

const (
	EnumDefaultGroceryList = -1
)

type GroceryList struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	GuildID   string    `json:"guild_id" gorm:"not null;index"`
	ListLabel string    `json:"list_label" gorm:"not null;index"`
	FancyName *string   `json:"fancy_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
