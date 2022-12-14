package telegram_client

import (
	"context"
	"os"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramClient struct {
}

func NewTelegramClient() *TelegramClient {
	return &TelegramClient{}
}

func (tc *TelegramClient) GetBotCommands(ctx context.Context) (res []tgBotApi.BotCommand, err error) {
	bot, err := tgBotApi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	res, err = bot.GetMyCommands()
	if err != nil {
		return nil, err
	}

	return

}
