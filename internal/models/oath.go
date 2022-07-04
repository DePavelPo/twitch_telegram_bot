package models

type TwitchOathResponse struct {
	Token     string `json:"access_token"`
	ExpiresIn uint64 `json:"expires_in"`
	TokenType string `json:"token_type"`
}
