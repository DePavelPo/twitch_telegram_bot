package telegram_service

import (
	"context"

	"twitch_telegram_bot/internal/models"
)

func (s *TelegramService) GetBotData(ctx context.Context) (res *models.TeleBotData, err error) {
	return s.telegramClient.GetData(ctx)
}
