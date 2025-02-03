package models

type Scope string

var (
	UserReadFollows Scope = "user:read:follows"
	ChannelReadSubs Scope = "channel:read:subscriptions"
)

// type TwitchOautGetTokenResponse struct {
// 	Token     string `json:"access_token"`
// 	ExpiresIn uint64 `json:"expires_in"`
// 	TokenType string `json:"token_type"`
// }

type TwitchOautValidateTokenResponse struct {
	ClientId  string   `json:"client_id"`
	Login     string   `json:"login"`
	Scopes    []string `json:"scopes"`
	UserId    string   `json:"user_id"`
	ExpiresIn uint64   `json:"expires_in"`
}

type TwitchOautGetTokenResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int32    `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"` // for user token
	Scope        []string `json:"scope"`         // for user token
	TokenType    string   `json:"token_type"`
}

type TokensWithChatID struct {
	AccessToken  *string `db:"access_token"`
	RefreshToken *string `db:"refresh_token"`
	ChatID       uint64  `db:"chat_id"`
}

type TokenWithState struct {
	AccessToken  *string `db:"access_token"`
	RefreshToken *string `db:"refresh_token"`
	State        string  `db:"current_state"`
}
