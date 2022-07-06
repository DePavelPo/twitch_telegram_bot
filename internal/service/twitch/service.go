package twitch_service

import twitch_client "twitch_telegram_bot/internal/client/twitch-client"

type TwitchService struct {
	twitchClient *twitch_client.TwitchClient
}

func NewService(twitchClient *twitch_client.TwitchClient) *TwitchService {
	return &TwitchService{
		twitchClient: twitchClient,
	}
}
