package twitch_user_authorization

import (
	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"

	dbRepository "twitch_telegram_bot/db/repository"
)

type TwitchUserAuthorizationService struct {
	dbRepo            *dbRepository.DBRepository
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
	protocol          string
	redirectAddr      string
}

func NewTwitchUserAuthorizationService(
	dbRepo *dbRepository.DBRepository,
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient,
	protocol, redirectAddr string,
) (*TwitchUserAuthorizationService, error) {
	return &TwitchUserAuthorizationService{
		dbRepo:            dbRepo,
		twitchOauthClient: twitchOauthClient,
		protocol:          protocol,
		redirectAddr:      redirectAddr,
	}, nil
}
