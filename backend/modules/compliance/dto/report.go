package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type ReportResponse struct {
	ID              uuid.UUID `json:"id"`
	FrameworkKey    string    `json:"frameworkKey"`
	FrameworkName   string    `json:"frameworkName"`
	FrameworkSource string    `json:"frameworkSource,omitempty"`
	GeneratedAt     time.Time `json:"generatedAt"`
	WindowFrom      time.Time `json:"windowFrom"`
	WindowTo        time.Time `json:"windowTo"`

	Summary  ReportSummary   `json:"summary"`
	Sections []ReportSection `json:"sections"`
	Controls []ControlRow    `json:"controls"`
}

type ReportBody struct {
	Summary  ReportSummary   `json:"summary"`
	Sections []ReportSection `json:"sections"`
	Controls []ControlRow    `json:"controls"`
}

// EditControlRequest annotates a control, and optionally overrides the engine
// on it.
//
// Status is optional because annotating and overriding are different acts. "We
// know, here is the remediation ticket" claims nothing about compliance, and
// requiring a status to record it is how a row ends up with a verdict nobody
// meant to give. The note is what is always required: an entry no one can
// explain is worth nothing to whoever reads the report later.
type EditControlRequest struct {
	Status domain.ComplianceStatus `json:"status,omitempty"`
	Note   string                  `json:"note" binding:"required,max=4000"`
}

type ScorePoint struct {
	Day         time.Time `json:"day"`
	GeneratedAt time.Time `json:"generatedAt"`
	Score       int       `json:"score"`
	Total       int       `json:"total"`
	Evaluated   int       `json:"evaluated"`
	Compliant   int       `json:"compliant"`
	HasDocument bool      `json:"hasDocument"`
}

type ReportMeta struct {
	ID            uuid.UUID `json:"id"`
	FrameworkKey  string    `json:"frameworkKey"`
	FrameworkName string    `json:"frameworkName"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Score         int       `json:"score"`
}

type ReportSummary struct {
	CompliantPct int `json:"compliantPct"`
	Total        int `json:"total"`
	Evaluated    int `json:"evaluated"`
	Compliant    int `json:"compliant"`
	NonCompliant int `json:"nonCompliant"`
	AtRisk       int `json:"atRisk"`
	NotCovered   int `json:"notCovered"`
	NotEvaluated int `json:"notEvaluated"`
	Pending      int `json:"pending"`
	OutOfScope   int `json:"outOfScope"`
}

type ReportSection struct {
	Key          string              `json:"key,omitempty"`
	Name         string              `json:"name"`
	Summary      ReportSummary       `json:"summary"`
	Requirements []ReportRequirement `json:"requirements"`
}

type ReportRequirement struct {
	ID         string                  `json:"id"`
	Name       string                  `json:"name"`
	Status     domain.ComplianceStatus `json:"status"`
	ControlIDs []string                `json:"controlIds"`
}

type ControlRow struct {
	ControlID      string                  `json:"controlId"`
	Name           string                  `json:"name"`
	Status         domain.ComplianceStatus `json:"status"`
	EngineStatus   domain.ComplianceStatus `json:"engineStatus"`
	Evidence       string                  `json:"evidence"` // one-line reason for the table; the detail is in Checks
	Coverage       int                     `json:"coverage"` // # enabled correlation rules covering this control
	Activity       int                     `json:"activity"` // # alerts from those rules in the window
	Checks         []CheckResult           `json:"checks,omitempty"`
	OriginalStatus domain.ComplianceStatus `json:"originalStatus,omitempty"`
	Note           string                  `json:"note,omitempty"`
	EditedBy       string                  `json:"editedBy,omitempty"`
	EditedAt       *time.Time              `json:"editedAt,omitempty"`
}

type CheckResult struct {
	Key      string              `json:"key"`
	Name     string              `json:"name"`
	Dataset  domain.Dataset      `json:"dataset,omitempty"`
	DataType string              `json:"dataType,omitempty"` // why a NOT_APPLICABLE check could not run
	Rule     domain.CheckRule    `json:"rule,omitempty"`
	Required *int                `json:"required,omitempty"` // the rule's threshold, so "0 of 1 required" reads on its own
	Outcome  domain.CheckOutcome `json:"outcome"`
	Hits     int64               `json:"hits"`
	Error    string              `json:"error,omitempty"`
}
