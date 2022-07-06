package telegram_service

import (
	telegreamClient "twitch_telegram_bot/internal/client/telegram-client"
)

type TelegramService struct {
	telegramClient *telegreamClient.TelegramClient
}

func NewService(telegramClient *telegreamClient.TelegramClient) *TelegramService {
	return &TelegramService{
		telegramClient: telegramClient,
	}
}
