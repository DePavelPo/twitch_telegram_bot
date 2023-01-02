package telegram_updates_check

import (
	"context"
	"fmt"
	"time"

	text_formater "twitch_telegram_bot/internal/utils/text-formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var exampleText string = `Request example:
	/twitch_user welovegames
	/twitch_user WELOVEGAMES
	`

func (tmcs *TelegramUpdatesCheckService) TwitchUserCase(ctx context.Context, msg tgbotapi.MessageConfig, updateInfo tgbotapi.Update) tgbotapi.MessageConfig {

	commandText := updateInfo.Message.Text[len(fmt.Sprint(twitchUserCommand)):]

	userLogin, isValid := validateText(commandText)
	if userLogin == "" || !isValid {
		msg.Text = invalidReq + exampleText
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return msg
	}

	userLogin = text_formater.ToLower(userLogin)

	users, err := tmcs.twitchClient.GetUserInfo(ctx, []string{userLogin})
	if err != nil {
		logrus.Error(err)
		msg.Text = somethingWrong
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return msg
	}

	if users == nil {
		msg.Text = "User not found"
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return msg
	}
	if len(users.Data) < 1 {
		msg.Text = "User not found"
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return msg
	}

	user := users.Data[0]
	accCreatedTime := user.CreatedAt

	// отображаем по МСК
	location := time.FixedZone("MSK", 3*60*60)
	accCreatedTime = accCreatedTime.In(location)

	var userType string
	switch user.Type {
	case "staff":
		userType = "twitch staff"
	case "admin":
		userType = "twitch admin"
	case "global_mod":
		userType = "global moderator"
	default:
		userType = "user"
	}

	var userBroadcasterType string
	switch user.BroadcasterType {
	case "partner":
		userBroadcasterType = "twitch partner"
	case "affiliate":
		userBroadcasterType = "twitch affiliate"
	default:
		userBroadcasterType = "user"
	}

	msg.Text = fmt.Sprintf(`
	User information:
	User: %s
	Account creation date: %s
	User type: %s
	Streamer type: %s
	`,
		user.DisplayName,
		accCreatedTime.Format("2006.01.02 15:04:05"),
		userType,
		userBroadcasterType)

	msg.ReplyToMessageID = updateInfo.Message.MessageID

	var streamStatus = "undefined"
	streams, err := tmcs.twitchClient.GetActiveStreamInfoByUsers(ctx, []string{userLogin})
	if err == nil || streams != nil {
		if len(streams.StreamInfo) < 1 {
			streamStatus = "offline"
		} else {
			stream := streams.StreamInfo[0]

			if stream.UserId == userLogin || stream.UserLogin == userLogin || stream.UserName == userLogin {
				streamStatus = "online"
			}
		}
	}

	msg.Text = fmt.Sprintf(`
	%s
	Stream status: %s
	%s
	`,
		msg.Text,
		streamStatus,
		fmt.Sprintf("https://www.twitch.tv/%s", user.Login))

	return msg
}
