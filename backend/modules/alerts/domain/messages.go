package domain

const (
	MsgStatusChangedNoObs   = "The user %s changed alert status from %s to %s"
	MsgStatusChangedWithObs = "The user %s changed alert status from %s to %s with the observation: %s"
	MsgStatusSystem         = "The system changed alert status from %s to %s with the observation: %s"
	MsgTagsManual           = "The user %s tagged the alert with: %s"
	MsgTagsManualCleared    = "The %s user removed all tags of the alert"
	MsgTagsAutomatic        = "The system tagged the alert automatically with: %s due a match with rule: %s"
	MsgNotesUpdated         = "The %s user updated the alert notes with: <br>%s"
	MsgNotesCleared         = "The %s user removed all notes of the alert"
	MsgConvertToIncident    = "This alert was added to incident %s by user %s"
	MsgAssigneeSet          = "The user %s assigned the alert to %s"
	MsgAssigneeCleared      = "The user %s unassigned the alert"
)
