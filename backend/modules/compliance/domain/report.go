package domain

// Control evaluation statuses (used in report rows and the rollup).
const (
	StatusCompliant    = "COMPLIANT"     // evaluated and passing
	StatusNonCompliant = "NON_COMPLIANT" // evaluated and failing
	StatusAtRisk       = "AT_RISK"       // passing analysis but weak coverage (or vice-versa)
	StatusNotCovered   = "NOT_COVERED"   // no checks and no covering rules
	StatusOutOfScope   = "OUT_OF_SCOPE"  // governance control — not SIEM-evaluable
	StatusPending      = "PENDING"       // data control whose check is not yet defined
)

type Report struct {
	FrameworkKey  string          `json:"frameworkKey"`
	FrameworkName string          `json:"frameworkName"`
	GeneratedAt   string          `json:"generatedAt"`
	Summary       ReportSummary   `json:"summary"`
	Sections      []ReportSection `json:"sections"`
}

type ReportSummary struct {
	CompliantPct int `json:"compliantPct"`
	Total        int `json:"total"`
	Compliant    int `json:"compliant"`
	NonCompliant int `json:"nonCompliant"`
	AtRisk       int `json:"atRisk"`
	NotCovered   int `json:"notCovered"`
	Pending      int `json:"pending"`
	OutOfScope   int `json:"outOfScope"`
}

type ReportSection struct {
	Name     string             `json:"name"`
	Controls []ReportControlRow `json:"controls"`
}

type ReportSnapshot struct {
	ID            string `json:"id"`
	TenantID      string `json:"-"`
	FrameworkKey  string `json:"frameworkKey"`
	FrameworkName string `json:"frameworkName"`
	Timestamp     string `json:"@timestamp"`
	Score         int    `json:"score"`
	Report        Report `json:"report"`
}

// ReportSnapshotMeta is the lightweight listing shape (no full report body).
type ReportSnapshotMeta struct {
	ID            string `json:"id"`
	FrameworkKey  string `json:"frameworkKey"`
	FrameworkName string `json:"frameworkName"`
	Timestamp     string `json:"@timestamp"`
	Score         int    `json:"score"`
}

type ReportControlRow struct {
	ControlID  string `json:"controlId"`
	Name       string `json:"name"`
	Status     string `json:"status"` // COMPLIANT | NON_COMPLIANT | AT_RISK | NOT_COVERED | OUT_OF_SCOPE | PENDING
	Evidence   string `json:"evidence"`
	Coverage   int    `json:"coverage"`             // # enabled correlation rules covering this control
	Activity   int    `json:"activity"`             // # alerts from those rules in the window
	Overridden bool   `json:"overridden,omitempty"` // true when status came from a manual override
	Note       string `json:"note,omitempty"`       // user note attached to this (framework, control)
}
