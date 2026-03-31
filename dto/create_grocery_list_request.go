package dto

// CreateGroceryListRequest is the JSON body for POST /grocery-lists. Clients must not send id, guild_id, or timestamps.
type CreateGroceryListRequest struct {
	ListLabel string `json:"list_label" validate:"required"`
	FancyName string `json:"fancy_name"`
}
