package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

type LoginRequest struct {
	// ProviderID names the directory to bind against when the user picked one.
	// Absent, every active directory of the tenant is tried, which is what a
	// single-directory install wants.
	ProviderID *uuid.UUID `json:"provider_id,omitempty"`
	Login      string     `json:"login" binding:"required" example:"admin"`
	Password   string     `json:"password" binding:"required" example:"changeme"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateMeRequest struct {
	Email   string `json:"email,omitempty" binding:"omitempty,email"`
	Name    string `json:"name,omitempty"`
	LangKey string `json:"lang_key,omitempty"`
}

type SessionResponse struct {
	ID        uuid.UUID `json:"id"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Current   bool      `json:"current"`
}

type UserResponse struct {
	ID         uuid.UUID         `json:"id"`
	Email      string            `json:"email"`
	Name       string            `json:"name,omitempty"`
	Status     domain.UserStatus `json:"status"`
	LangKey    string            `json:"lang_key,omitempty"`
	ImageURL   string            `json:"image_url,omitempty"`
	Federated  bool              `json:"federated"`
	TfaEnabled bool              `json:"tfa_enabled"`
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type" example:"Bearer"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type LoginResponse struct {
	TokenPair
	User        UserResponse `json:"user"`
	TfaRequired bool         `json:"tfa_required,omitempty"`
	// Which factor the user must produce, so the screen can say "check your
	// mail" or "open your authenticator" instead of guessing.
	TfaType      domain.TfaFactorType `json:"tfa_type,omitempty"`
	PreAuthToken string               `json:"pre_auth_token,omitempty"`
}

func ToUserResponse(u domain.User, tfaEnabled bool) UserResponse {
	return UserResponse{
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		Status:     u.Status,
		LangKey:    u.LangKey,
		ImageURL:   u.ImageURL,
		Federated:  u.IdentityProviderID != nil,
		TfaEnabled: tfaEnabled,
	}
}
