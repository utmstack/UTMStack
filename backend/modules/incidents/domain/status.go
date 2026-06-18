package domain

type IncidentStatus string

const (
	StatusOpen      IncidentStatus = "OPEN"
	StatusInReview  IncidentStatus = "IN_REVIEW"
	StatusCompleted IncidentStatus = "COMPLETED"
	StatusMerged    IncidentStatus = "MERGED"
)

func (s IncidentStatus) ToAlertStatus() int {
	switch s {
	case StatusOpen:
		return 2
	case StatusInReview:
		return 3
	case StatusCompleted:
		return 5
	case StatusMerged:
		return 0
	default:
		return 2
	}
}

func (s IncidentStatus) Label() string {
	switch s {
	case StatusOpen:
		return "Open"
	case StatusInReview:
		return "In review"
	case StatusCompleted:
		return "Completed"
	case StatusMerged:
		return "Merged"
	default:
		return string(s)
	}
}

func AlertStatusLabel(status int) string {
	switch status {
	case 0:
		return "Merged"
	case 2:
		return "Open"
	case 3:
		return "In review"
	case 5:
		return "Completed"
	default:
		return "Unknown"
	}
}

func (s IncidentStatus) Valid() bool {
	switch s {
	case StatusOpen, StatusInReview, StatusCompleted, StatusMerged:
		return true
	}
	return false
}
