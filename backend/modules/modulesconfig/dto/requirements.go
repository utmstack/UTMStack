package dto

import "github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"

type CheckRequirementsResponse struct {
	Status domain.ModuleRequirementStatus `json:"status"`
	Checks []domain.ModuleRequirement     `json:"checks"`
}
