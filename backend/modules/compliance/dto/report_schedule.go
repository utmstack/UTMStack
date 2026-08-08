package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateScheduleRequest struct {
	FrameworkKey   string `json:"frameworkKey" binding:"required"`
	ScheduleString string `json:"scheduleString" binding:"required,max=250"`
	WindowDays     int    `json:"windowDays" binding:"omitempty,min=1,max=3650"`
	To             string `json:"to" binding:"required,max=2000"`
	Cc             string `json:"cc" binding:"max=2000"`
}

type UpdateScheduleRequest struct {
	ID             uuid.UUID `json:"id" binding:"required"`
	FrameworkKey   string    `json:"frameworkKey" binding:"required"`
	ScheduleString string    `json:"scheduleString" binding:"required,max=250"`
	WindowDays     int       `json:"windowDays" binding:"omitempty,min=1,max=3650"`
	To             string    `json:"to" binding:"required,max=2000"`
	Cc             string    `json:"cc" binding:"max=2000"`
}

type ScheduleResponse struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"userId"`
	FrameworkKey      string    `json:"frameworkKey"`
	ScheduleString    string    `json:"scheduleString"`
	WindowDays        int       `json:"windowDays"`
	To                string    `json:"to"`
	Cc                string    `json:"cc"`
	LastExecutionDate time.Time `json:"lastExecutionDate"`
}

type ScheduleFilters struct {
	Page         int    `form:"page"`
	Size         int    `form:"size"`
	FrameworkKey string `form:"frameworkKey"`
}
