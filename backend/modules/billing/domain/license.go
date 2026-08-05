package domain

import "time"

type Edition string

const (
	EditionCommunity  Edition = "community"
	EditionEnterprise Edition = "enterprise"
)

type License struct {
	Edition          Edition   `json:"edition"`
	MSSP             bool      `json:"mssp"`
	IngestGBPerMonth int64     `json:"ingestGbPerMonth"` // contracted volume; 0 = unlimited
	Type             string    `json:"type"`             // online | offline ("" when community)
	ExpiresAt        time.Time `json:"expiresAt,omitempty"`
}

func Community() License { return License{Edition: EditionCommunity} }

func (l License) IsEnterprise() bool { return l.Edition == EditionEnterprise }
func (l License) IsMSSP() bool       { return l.MSSP }
