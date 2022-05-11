package dto

import "github.com/verzac/grocer-discord-bot/models"

type GuildGroceryList struct {
	GuildID        string                `json:"guild_id"`
	GroceryEntries []models.GroceryEntry `json:"grocery_entries"`
	GroceryLists   []models.GroceryList  `json:"grocery_lists"`
}
