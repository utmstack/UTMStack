package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

type LoginRequest struct {
	Login    string `json:"login" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"changeme"`
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

type ResetPasswordInitRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordFinishRequest struct {
	Key         string `json:"key" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type UpdateMeRequest struct {
	Email     string `json:"email,omitempty" binding:"omitempty,email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	LangKey   string `json:"lang_key,omitempty"`
}

type SessionResponse struct {
	ID        uint64    `json:"id"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Current   bool      `json:"current"`
}

type UserResponse struct {
	ID        uint64 `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Activated bool   `json:"activated"`
	LangKey   string `json:"lang_key,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
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
	User         UserResponse `json:"user"`
	TfaRequired  bool         `json:"tfa_required,omitempty"`
	TfaMethod    string       `json:"tfa_method,omitempty"`
	PreAuthToken string       `json:"pre_auth_token,omitempty"`
}

func ToUserResponse(u domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Login:     u.Login,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Activated: u.Activated,
		LangKey:   u.LangKey,
		ImageURL:  u.ImageURL,
	}
}
