package twitch_service

import (
	"context"
	"twitch_telegram_bot/internal/models"
)

func (tws *TwitchService) GetOAuthToken(ctx context.Context) (*models.TwitchOatGetTokenhResponse, error) {
	return tws.twitchClient.TwitchOAuthGetToken(ctx)
}
