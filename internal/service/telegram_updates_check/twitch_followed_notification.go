package telegram_updates_check

import (
	"context"
	"twitch_telegram_bot/internal/models"
	formater "twitch_telegram_bot/internal/utils/formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (tmcs *TelegramUpdatesCheckService) twitchAddFollowedNotification(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (msg tgbotapi.MessageConfig, err error) {

	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	data, err := tmcs.twitchUserAuthservice.CheckUserTokensByChat(ctx, uint64(updateInfo.Message.Chat.ID))
	if err != nil {

		msg.Text = somethingWrong

		return msg, errors.Wrap(err, "CheckUserTokensByChat")
	}

	if data.Link != "" {

		resp := "To use this feature, follow the link and provide access to the necessary information"
		msg.Text = resp

		msg = formater.CreateTelegramSingleButtonLink(msg, data.Link, "Open Link", updateInfo.Message.MessageID)

		return
	}

	err = tmcs.dbRepo.AddTwitchNotification(ctx, uint64(updateInfo.Message.Chat.ID), data.UserID, models.NotificationFollowed)
	if err != nil {
		msg.Text = somethingWrong

		return msg, errors.Wrap(err, "AddTwitchNotification")
	}

	msg.Text = "Request successfully accepted! This channel will now receive stream notifications from channels that you're following"

	return
}

func (tmcs *TelegramUpdatesCheckService) twitchCancelFollowedNotification(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (msg tgbotapi.MessageConfig, err error) {

	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	err = tmcs.dbRepo.SetInactiveNotificationByType(ctx, uint64(updateInfo.Message.Chat.ID), "", models.NotificationFollowed)
	if err != nil {
		if err.Error() == "notification not found" {
			msg.Text = "No requests for notifications were found 😕"

			return msg, errors.Errorf("followed streams notification by chatId %d not found", updateInfo.Message.Chat.ID)
		}
		msg.Text = somethingWrong

		return msg, errors.Wrap(err, "SetInactiveNotificationByType")
	}

	msg.Text = "Notifications were disabled successfully"

	return
}
