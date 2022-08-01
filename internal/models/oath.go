package models

type TwitchOautGetTokenResponse struct {
	Token     string `json:"access_token"`
	ExpiresIn uint64 `json:"expires_in"`
	TokenType string `json:"token_type"`
}

type TwitchOautValidateTokenResponse struct {
	ClientId  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserId    string   `json:"user_id"`
	ExpiresIn uint64   `json:"expires_in"`
}
