package models

const (
	TokenInvalid        string = "token invalid"
	RefreshTokenInvalid string = "Invalid refresh token"
)

type GetUserUnauthorized struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type ValidateTokenInvalid struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
