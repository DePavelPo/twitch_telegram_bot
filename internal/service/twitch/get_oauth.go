package twitch_service

import (
	"context"
	"twitch_telegram_bot/internal/models"
)

func (tws *TwitchService) GetOAuthToken(ctx context.Context) (*models.TwitchOathResponse, error) {
	return tws.twitchClient.TwitchOAuth(ctx)
}
