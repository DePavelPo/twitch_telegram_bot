package telegram

import (
	teleService "twitch_telegram_bot/internal/service/telegram"
)

type TelegramHandler struct {
	telegramService *teleService.TelegramService
}

func NewTelegramHandler(telegramService *teleService.TelegramService) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
	}
}
