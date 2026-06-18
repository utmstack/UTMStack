package domain

type NotificationSource string

const (
	SourceAS400        NotificationSource = "AS400"
	SourceBitDefender  NotificationSource = "BIT_DEFENDER"
	SourceAWS          NotificationSource = "AWS"
	SourceAzure        NotificationSource = "AZURE"
	SourceOffice365    NotificationSource = "OFFICE_365"
	SourceSophos       NotificationSource = "SOPHOS"
	SourceGoogle       NotificationSource = "GOOGLE"
	SourceEmailSetting NotificationSource = "EMAIL_SETTING"
	SourceAlerts       NotificationSource = "ALERTS"
	SourceSystem       NotificationSource = "SYSTEM"
)

func (s NotificationSource) Valid() bool {
	switch s {
	case SourceAS400, SourceBitDefender, SourceAWS, SourceAzure,
		SourceOffice365, SourceSophos, SourceGoogle, SourceEmailSetting, SourceAlerts:
		return true
	}
	return false
}

type NotificationType string

const (
	TypeInfo    NotificationType = "INFO"
	TypeWarning NotificationType = "WARNING"
	TypeError   NotificationType = "ERROR"
)

func (t NotificationType) Valid() bool {
	switch t {
	case TypeInfo, TypeWarning, TypeError:
		return true
	}
	return false
}

type NotificationStatus string

const (
	StatusActive  NotificationStatus = "ACTIVE"
	StatusHidden  NotificationStatus = "HIDDEN"
	StatusDeleted NotificationStatus = "DELETED"
)

func (s NotificationStatus) Valid() bool {
	switch s {
	case StatusActive, StatusHidden, StatusDeleted:
		return true
	}
	return false
}
