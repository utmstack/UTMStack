package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

// CheckRequirementsResponse aggregates a ModuleKind's preflight checks. If any
// check failed the top-level Status is FAIL, matching the legacy panel
// expectation.
type CheckRequirementsResponse struct {
	Status domain.ModuleRequirementStatus `json:"status"`
	Checks []domain.ModuleRequirement     `json:"checks"`
}
