package dto

type CreateGroceryListRequest struct {
	ListLabel string  `json:"list_label" validate:"required"`
	FancyName *string `json:"fancy_name"`
}

type UpdateGroceryListRequest struct {
	FancyName *string `json:"fancy_name" validate:"required"`
}
