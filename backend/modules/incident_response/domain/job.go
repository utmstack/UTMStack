package domain

import "time"

type JobStatus int

const (
	JobStatusPending  JobStatus = 0
	JobStatusRunning  JobStatus = 1
	JobStatusExecuted JobStatus = 2
	JobStatusError    JobStatus = 3
)

func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "PENDING"
	case JobStatusRunning:
		return "RUNNING"
	case JobStatusExecuted:
		return "EXECUTED"
	default:
		return "ERROR"
	}
}

type UtmIncidentJob struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ActionID     *int64     `gorm:"column:action_id"                  json:"actionId"`
	Params       *string    `gorm:"column:params"                     json:"params"`
	Agent        *string    `gorm:"column:agent"                      json:"agent"`
	Status       *int       `gorm:"column:status"                     json:"status"`
	JobResult    *string    `gorm:"column:job_result"                 json:"jobResult"`
	OriginID     int        `gorm:"column:origin_id;not null"         json:"originId"`
	OriginType   string     `gorm:"column:origin_type;not null"       json:"originType"`
	CreatedDate  time.Time  `gorm:"column:created_date;not null"      json:"createdDate"`
	ModifiedDate *time.Time `gorm:"column:modified_date"              json:"modifiedDate"`
	CreatedUser  string     `gorm:"column:created_user;not null"      json:"createdUser"`
	ModifiedUser *string    `gorm:"column:modified_user"              json:"modifiedUser"`
}

func (UtmIncidentJob) TableName() string { return "utm_incident_jobs" }
