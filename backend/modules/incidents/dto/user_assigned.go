package dto

import "github.com/google/uuid"

type UserAssignedDTO struct {
	ID    uuid.UUID `json:"id"`
	Login string    `json:"login"`
}
