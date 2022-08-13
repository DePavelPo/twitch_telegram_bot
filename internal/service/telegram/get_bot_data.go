package telegram_service

import (
	"context"
	"fmt"

	"twitch_telegram_bot/internal/models"
)

func (s *TelegramService) GetBotCommands(ctx context.Context) (res *models.TeleBotCommands, err error) {

	data, err := s.telegramClient.GetBotCommands(ctx)
	if err != nil {
		return nil, err
	}

	res = &models.TeleBotCommands{}

	for _, commandInfo := range data {

		command := models.TeleBotCommand{
			Command:     fmt.Sprintf("/%s", commandInfo.Command),
			Description: commandInfo.Description,
		}

		res.Commands = append(res.Commands, command)
	}

	return
}
