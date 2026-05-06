package dto

type GroceryBatchDeleteRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1,max=300,dive,gt=0"`
}
