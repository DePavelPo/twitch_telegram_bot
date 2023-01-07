package twitch_user_authorization

import (
	"github.com/jmoiron/sqlx"

	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"
)

type TwitchUserAuthorizationService struct {
	db                *sqlx.DB
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

func NewTwitchUserAuthorizationService(db *sqlx.DB, twitchOauthClient *twitch_oauth_client.TwitchOauthClient) (*TwitchUserAuthorizationService, error) {
	return &TwitchUserAuthorizationService{
		db:                db,
		twitchOauthClient: twitchOauthClient,
	}, nil
}
