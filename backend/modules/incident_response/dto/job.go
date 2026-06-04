package dto

import "time"

type CreateJobRequest struct {
	ActionID   *int64  `json:"actionId"`
	Params     *string `json:"params"`
	Agent      *string `json:"agent"`
	Status     *int    `json:"status"`
	OriginID   int     `json:"originId"`
	OriginType string  `json:"originType"`
}

type JobResponse struct {
	ID           int64      `json:"id"`
	ActionID     *int64     `json:"actionId"`
	Params       *string    `json:"params"`
	Agent        *string    `json:"agent"`
	Status       *int       `json:"status"`
	JobResult    *string    `json:"jobResult"`
	OriginID     int        `json:"originId"`
	OriginType   string     `json:"originType"`
	CreatedDate  time.Time  `json:"createdDate"`
	ModifiedDate *time.Time `json:"modifiedDate"`
	CreatedUser  string     `json:"createdUser"`
	ModifiedUser *string    `json:"modifiedUser"`
}
