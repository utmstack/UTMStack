package dto

import "github.com/utmstack/utmstack/backend/modules/alerts/domain"

type AdversaryResponse struct {
	Adversary *domain.Side        `json:"adversary"`
	Alerts    []AlertWithChildren `json:"alerts"`
}

type AlertWithChildren struct {
	Alert    domain.UtmAlert   `json:"alert"`
	Children []domain.UtmAlert `json:"children"`
}

type AdversaryWrapper struct {
	Adversary *domain.Side `json:"adversary"`
}
