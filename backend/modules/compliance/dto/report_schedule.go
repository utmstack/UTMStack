package dto

import "time"

type CreateScheduleRequest struct {
	FrameworkKey   string `json:"frameworkKey" binding:"required"`
	ScheduleString string `json:"scheduleString" binding:"required,max=250"`
	Recipients     string `json:"recipients"`
}

type UpdateScheduleRequest struct {
	ID             int64  `json:"id" binding:"required"`
	FrameworkKey   string `json:"frameworkKey" binding:"required"`
	ScheduleString string `json:"scheduleString" binding:"required,max=250"`
	Recipients     string `json:"recipients"`
}

type ScheduleResponse struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"userId"`
	FrameworkKey      string    `json:"frameworkKey"`
	ScheduleString    string    `json:"scheduleString"`
	Recipients        string    `json:"recipients"`
	LastExecutionDate time.Time `json:"lastExecutionDate"`
}

type ScheduleFilters struct {
	Page         int    `form:"page"`
	Size         int    `form:"size"`
	FrameworkKey string `form:"frameworkKey"`
}
