package telegram_client

import (
	"context"
	"os"

	"twitch_telegram_bot/internal/models"

	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramClient struct {
}

func NewTelegramClient() *TelegramClient {
	return &TelegramClient{}
}

func (tc *TelegramClient) GetData(ctx context.Context) (*models.TeleBotData, error) {
	bot, err := tgBotApi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	botCommands, err := bot.GetMyCommands()
	if err != nil {
		return nil, err
	}

	var res models.TeleBotData

	for _, botCommand := range botCommands {
		res.Commands = append(res.Commands, botCommand.Command)
	}

	return &res, nil

}
