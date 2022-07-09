package models

type GetUserUnauthorized struct {
	Error   string `json:"error"`
	Status  int `json:"status"`
	Message string `json:"message"`
}
