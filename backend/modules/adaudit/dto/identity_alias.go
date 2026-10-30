package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
)

type CreateIdentityRequest struct {
	Name string `json:"name"`
}

type IdentityResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type IdentityListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Count     int       `json:"count"`
}

type IdentityDetail struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	Users     []UserRow `json:"users"`
}

type UserRow struct {
	domain.ADUser
	Identity *string `json:"identity"`
}

type AssignIdentityRequest struct {
	IdentityID string `json:"identityId"`
}
