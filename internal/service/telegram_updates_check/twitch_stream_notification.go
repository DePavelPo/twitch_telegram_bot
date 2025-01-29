package telegram_updates_check

import (
	"context"
	"fmt"
	"twitch_telegram_bot/internal/models"
	formater "twitch_telegram_bot/internal/utils/formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (tmcs *TelegramUpdatesCheckService) TwitchAddStreamNotification(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (msg tgbotapi.MessageConfig, err error) {

	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	commandText := updateInfo.Message.Text[len(fmt.Sprint(twitchStreamNotifi)):]

	userLogin, isValid := validateText(commandText)
	if !isValid {
		msg.Text = invalidReq + fmt.Sprintf(userCustomExampleText, twitchStreamNotifi, twitchStreamNotifi)

		return
	}

	userLoginLowercase := formater.ToLower(userLogin)

	err = tmcs.dbRepo.AddTwitchNotification(ctx, uint64(updateInfo.Message.Chat.ID), userLoginLowercase, models.NotificationByUser)
	if err != nil {
		msg.Text = somethingWrong

		return msg, errors.Wrap(err, "AddTwitchNotification")
	}

	msg.Text = fmt.Sprintf("Request successfully accepted! This channel will now receive stream notifications from %s Twitch channel", userLogin)

	return
}
