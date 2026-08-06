package dto

type ResetPasswordInitRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordFinishRequest struct {
	Key         string `json:"key" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
