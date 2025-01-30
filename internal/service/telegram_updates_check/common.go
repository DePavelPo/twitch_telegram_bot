package telegram_updates_check

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (tmcs *TelegramUpdatesCheckService) start(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (msg tgbotapi.MessageConfig, err error) {

	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	msg.Text = "Greetings! The bot provides the functionality for interacting with Twitch streaming platform\n"

	teleCommands, err := tmcs.telegramService.GetBotCommands(ctx)
	if err != nil {

		return msg, errors.Wrap(err, "GetBotCommands")
	}

	if teleCommands != nil {
		msg.Text = fmt.Sprintf("%s\n%s", msg.Text, "Bot's command list:")

		for _, teleCommand := range teleCommands.Commands {
			msg.Text = fmt.Sprintf("%s\n%s - %s", msg.Text, teleCommand.Command, teleCommand.Description)
		}

	}

	return
}

func (tmcs *TelegramUpdatesCheckService) commands(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (msg tgbotapi.MessageConfig, err error) {

	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	teleCommands, err := tmcs.telegramService.GetBotCommands(ctx)
	if err != nil {

		msg.Text = somethingWrong

		return msg, errors.Wrap(err, "GetBotCommands")
	}

	if teleCommands != nil {

		msg.Text = "Bot's command list:\n"

		for _, teleCommand := range teleCommands.Commands {
			msg.Text = fmt.Sprintf("%s\n%s - %s", msg.Text, teleCommand.Command, teleCommand.Description)
		}

	}

	return
}
