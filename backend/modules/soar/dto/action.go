package dto

import "time"

type CreateActionRequest struct {
	ActionCommand     *string `json:"actionCommand"`
	ActionDescription *string `json:"actionDescription"`
	ActionParams      *string `json:"actionParams"`
	ActionType        *int    `json:"actionType"`
	ActionEditable    bool    `json:"actionEditable"`
}

type UpdateActionRequest struct {
	ID                int64   `json:"id"  binding:"required"`
	ActionCommand     *string `json:"actionCommand"`
	ActionDescription *string `json:"actionDescription"`
	ActionParams      *string `json:"actionParams"`
	ActionType        *int    `json:"actionType"`
	ActionEditable    bool    `json:"actionEditable"`
}

type ActionResponse struct {
	ID                int64      `json:"id"`
	ActionCommand     *string    `json:"actionCommand"`
	ActionDescription *string    `json:"actionDescription"`
	ActionParams      *string    `json:"actionParams"`
	ActionType        *int       `json:"actionType"`
	ActionEditable    bool       `json:"actionEditable"`
	CreatedDate       time.Time  `json:"createdDate"`
	ModifiedDate      *time.Time `json:"modifiedDate"`
	CreatedUser       string     `json:"createdUser"`
	ModifiedUser      *string    `json:"modifiedUser"`
}

type ActionFilter struct {
	Page           int     `form:"page"`
	Size           int     `form:"size"`
	ActionCommand  *string `form:"actionCommand"`
	ActionType     *int    `form:"actionType"`
	ActionEditable *bool   `form:"actionEditable"`
}
