package usecase

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

func windowOf(now time.Time, days int) (from, to time.Time) {
	if days <= 0 {
		days = domain.DefaultWindowDays
	}
	return now.AddDate(0, 0, -days), now
}

func validStatus(s domain.ComplianceStatus) bool {
	switch s {
	case domain.StatusCompliant, domain.StatusNonCompliant, domain.StatusAtRisk,
		domain.StatusNotCovered, domain.StatusNotEvaluated, domain.StatusPending,
		domain.StatusOutOfScope:
		return true
	}
	return false
}

func scored(s domain.ComplianceStatus) bool {
	switch s {
	case domain.StatusOutOfScope, domain.StatusPending, domain.StatusNotEvaluated:
		return false
	}
	return true
}
func statusRank(s domain.ComplianceStatus) int {
	switch s {
	case domain.StatusNonCompliant:
		return 6
	case domain.StatusNotCovered:
		return 5
	case domain.StatusAtRisk:
		return 4
	case domain.StatusCompliant:
		return 3
	case domain.StatusNotEvaluated:
		return 2
	case domain.StatusPending:
		return 1
	default: // StatusOutOfScope
		return 0
	}
}

func rollUp(parts []domain.ComplianceStatus) domain.ComplianceStatus {
	if len(parts) == 0 {
		return domain.StatusNotCovered
	}
	worst := parts[0]
	for _, p := range parts[1:] {
		if statusRank(p) > statusRank(worst) {
			worst = p
		}
	}
	return worst
}

func addToSummary(s *dto.ReportSummary, status domain.ComplianceStatus) {
	s.Total++
	if scored(status) {
		s.Evaluated++
	}
	switch status {
	case domain.StatusCompliant:
		s.Compliant++
	case domain.StatusNonCompliant:
		s.NonCompliant++
	case domain.StatusAtRisk:
		s.AtRisk++
	case domain.StatusNotCovered:
		s.NotCovered++
	case domain.StatusNotEvaluated:
		s.NotEvaluated++
	case domain.StatusPending:
		s.Pending++
	case domain.StatusOutOfScope:
		s.OutOfScope++
	}
}

func finalize(s *dto.ReportSummary) {
	if s.Total == 0 {
		s.CompliantPct = 0
		return
	}
	s.CompliantPct = (s.Compliant * 100) / s.Total
}

func sectionsOf(fw *domain.Framework, controls []dto.ControlRow) []dto.ReportSection {
	out := make([]dto.ReportSection, 0, len(fw.Sections))
	for _, sec := range fw.Sections {
		rs := dto.ReportSection{
			Key:          sec.Key,
			Name:         sec.Name,
			Requirements: make([]dto.ReportRequirement, 0, len(sec.Requirements)),
		}
		for _, req := range sec.Requirements {
			rs.Requirements = append(rs.Requirements, dto.ReportRequirement{
				ID:         req.ID,
				Name:       req.Name,
				ControlIDs: req.SatisfiedBy,
			})
		}
		out = append(out, rs)
	}
	return out
}

func recompute(body *dto.ReportBody) {
	byID := make(map[string]domain.ComplianceStatus, len(body.Controls))
	for _, c := range body.Controls {
		byID[c.ControlID] = c.Status
	}

	body.Summary = dto.ReportSummary{}
	for i := range body.Sections {
		sec := &body.Sections[i]
		sec.Summary = dto.ReportSummary{}
		for j := range sec.Requirements {
			req := &sec.Requirements[j]
			parts := make([]domain.ComplianceStatus, 0, len(req.ControlIDs))
			for _, cid := range req.ControlIDs {
				if st, ok := byID[cid]; ok {
					parts = append(parts, st)
				}
			}
			req.Status = rollUp(parts)
			addToSummary(&sec.Summary, req.Status)
			addToSummary(&body.Summary, req.Status)
		}
		finalize(&sec.Summary)
	}
	finalize(&body.Summary)
}

func inheritEdit(row *dto.ControlRow, prev []dto.ControlRow) {
	for i := range prev {
		p := &prev[i]
		if p.ControlID != row.ControlID || p.EditedAt == nil {
			continue
		}
		// The note always survives: a remediation ticket is as relevant on the
		// run that finally passes as on the one that failed.
		row.Note = p.Note
		row.EditedBy = p.EditedBy
		row.EditedAt = p.EditedAt
		// The verdict survives only if there was one. An annotation leaves the
		// engine's status in place, so the row keeps tracking the engine while
		// still carrying what a person wrote.
		if p.OriginalStatus != "" {
			row.Status = p.Status
			row.OriginalStatus = p.OriginalStatus
		}
		return
	}
}

func effectiveDataset(ch domain.Check) domain.Dataset {
	if ch.Dataset == domain.DatasetAlerts {
		return domain.DatasetAlerts
	}
	return domain.DatasetLogs
}

func validRule(r domain.CheckRule) bool {
	return r == domain.RuleMinHitsRequired || r == domain.RuleThresholdMax
}

func checkRunnable(ch domain.Check) bool { return !ch.Todo && validRule(ch.Rule) }

func effectiveScope(c *domain.Control) domain.ControlScope {
	if c.Scope == domain.ScopeGovernance {
		return domain.ScopeGovernance
	}
	return domain.ScopeData
}

func effectiveStrategy(c *domain.Control) domain.CheckStrategy {
	if c.Strategy == domain.StrategyAny {
		return domain.StrategyAny
	}
	return domain.StrategyAll
}

func runnableChecks(c *domain.Control) []domain.Check {
	out := make([]domain.Check, 0, len(c.Checks))
	for _, ch := range c.Checks {
		if checkRunnable(ch) {
			out = append(out, ch)
		}
	}
	return out
}
