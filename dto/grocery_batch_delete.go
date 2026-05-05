package dto

type GroceryBatchDeleteRequest struct {
	IDs []int64 `json:"ids" validate:"required,min=1,max=300,dive,gt=0"`
}
