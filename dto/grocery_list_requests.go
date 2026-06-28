package dto

type CreateGroceryListRequest struct {
	ListLabel string  `json:"list_label" validate:"required"`
	FancyName *string `json:"fancy_name"`
}
