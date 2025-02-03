package twitch_oath_client

import (
	"net/http"
)

const (
	twitchIDSchemeHost string = "https://id.twitch.tv"
)

var expectableErrorCode = map[int]bool{
	http.StatusBadRequest:   true,
	http.StatusUnauthorized: true,
	http.StatusForbidden:    true,
}

type TwitchOauthClient struct {
	protocol     string
	redirectAddr string
}

func NewTwitchOauthClient(
	protocol, redirectAddr string,
) *TwitchOauthClient {
	return &TwitchOauthClient{
		protocol:     protocol,
		redirectAddr: redirectAddr,
	}
}
