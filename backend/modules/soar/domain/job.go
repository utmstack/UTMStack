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

// UtmIncidentJob records an incident-response command run against an agent.
// Reports surface these as the "responses" executed during incident handling.
//
// origin_id is varchar(100) in the legacy schema (Hibernate stored an Integer
// into it); it is kept as a string here so an in-place AutoMigrate over the
// legacy table does not attempt an illegal varchar->bigint column cast.
type UtmIncidentJob struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement"   json:"id"`
	ActionID     *int64     `gorm:"column:action_id"                     json:"actionId"`
	Params       *string    `gorm:"column:params"                        json:"params"`
	Agent        *string    `gorm:"column:agent"                         json:"agent"`
	Status       *int       `gorm:"column:status"                        json:"status"`
	JobResult    *string    `gorm:"column:job_result"                    json:"jobResult"`
	OriginID     string     `gorm:"column:origin_id;size:100;not null"   json:"originId"`
	OriginType   string     `gorm:"column:origin_type;size:30;not null"  json:"originType"`
	CreatedDate  time.Time  `gorm:"column:created_date;not null"         json:"createdDate"`
	ModifiedDate *time.Time `gorm:"column:modified_date"                 json:"modifiedDate"`
	CreatedUser  string     `gorm:"column:created_user;size:50;not null" json:"createdUser"`
	ModifiedUser *string    `gorm:"column:modified_user;size:50"         json:"modifiedUser"`
}

func (UtmIncidentJob) TableName() string { return "utm_incident_jobs" }
