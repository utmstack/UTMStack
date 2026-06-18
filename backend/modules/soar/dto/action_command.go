package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type CreateActionCommandRequest struct {
	ActionID   int64   `json:"actionId"   binding:"required"`
	OsPlatform *string `json:"osPlatform"`
	Command    *string `json:"command"`
}

type UpdateActionCommandRequest struct {
	ID         int64   `json:"id"         binding:"required"`
	ActionID   int64   `json:"actionId"   binding:"required"`
	OsPlatform *string `json:"osPlatform"`
	Command    *string `json:"command"`
}

type ActionCommandResponse struct {
	ID         int64   `json:"id"`
	ActionID   int64   `json:"actionId"`
	OsPlatform *string `json:"osPlatform"`
	Command    *string `json:"command"`
}

type ActionCommandFilter struct {
	database.Params
	ActionID   *int64  `form:"actionId"`
	OsPlatform *string `form:"osPlatform"`
	Command    *string `form:"command"`
}
