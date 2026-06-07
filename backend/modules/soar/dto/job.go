package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/pkg/database"
)

type CreateJobRequest struct {
	ActionID   *int64  `json:"actionId"`
	Params     *string `json:"params"`
	Agent      *string `json:"agent"`
	Status     *int    `json:"status"`
	OriginID   string  `json:"originId"`
	OriginType string  `json:"originType"`
}

type JobResponse struct {
	ID           int64      `json:"id"`
	ActionID     *int64     `json:"actionId"`
	Params       *string    `json:"params"`
	Agent        *string    `json:"agent"`
	Status       *int       `json:"status"`
	JobResult    *string    `json:"jobResult"`
	OriginID     string     `json:"originId"`
	OriginType   string     `json:"originType"`
	CreatedDate  time.Time  `json:"createdDate"`
	ModifiedDate *time.Time `json:"modifiedDate"`
	CreatedUser  string     `json:"createdUser"`
	ModifiedUser *string    `json:"modifiedUser"`
}

type JobFilter struct {
	database.Params
	ActionID   *int64  `form:"actionId"`
	Agent      *string `form:"agent"`
	Status     *int    `form:"status"`
	OriginID   *int    `form:"originId"`
	OriginType *string `form:"originType"`
}
