package twitch_service

import (
	twitch_client "twitch_telegram_bot/internal/client/twitch-client"
	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"
)

type TwitchService struct {
	twitchClient      *twitch_client.TwitchClient
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

func NewService(twitchClient *twitch_client.TwitchClient, twitchOauthClient *twitch_oauth_client.TwitchOauthClient) *TwitchService {
	return &TwitchService{
		twitchClient:      twitchClient,
		twitchOauthClient: twitchOauthClient,
	}
}
