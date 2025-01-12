package telegram_updates_check

import (
	"context"
	"fmt"
	"time"

	formater "twitch_telegram_bot/internal/utils/formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var userCustomExampleText string = `Request example:
	%s welovegames
	%s WELOVEGAMES
	`

func (tmcs *TelegramUpdatesCheckService) TwitchUserCase(
	ctx context.Context,
	photo tgbotapi.PhotoConfig,
	updateInfo tgbotapi.Update,
	chatID int64,
) tgbotapi.PhotoConfig {

	commandText := updateInfo.Message.Text[len(fmt.Sprint(twitchUserCommand)):]

	userLogin, isValid := validateText(commandText)
	if userLogin == "" || !isValid {
		photo.Caption = invalidReq + fmt.Sprintf(userCustomExampleText, twitchUserCommand, twitchUserCommand)
		photo.ReplyToMessageID = updateInfo.Message.MessageID
		return photo
	}

	userLogin = formater.ToLower(userLogin)

	users, err := tmcs.twitchClient.GetUserInfo(ctx, []string{userLogin})
	if err != nil {
		logrus.Error(err)
		photo.Caption = somethingWrong
		photo.ReplyToMessageID = updateInfo.Message.MessageID
		return photo
	}

	if (users == nil) || (len(users.Data) < 1) {
		photo.Caption = "User not found"
		photo.ReplyToMessageID = updateInfo.Message.MessageID
		return photo
	}

	user := users.Data[0]
	accCreatedTime := user.CreatedAt

	// using MSK timezone
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

	photo.Caption = fmt.Sprintf(`
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

	photo.ReplyToMessageID = updateInfo.Message.MessageID

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

	photo.Caption = fmt.Sprintf(`
	%s
	Stream status: %s
	`,
		photo.Caption,
		streamStatus,
	)

	newPhoto := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(user.ProfileImageUrl))
	newPhoto.ReplyToMessageID = updateInfo.Message.MessageID
	newPhoto.Caption = photo.Caption

	newPhoto = formater.CreateTelegramSingleButtonLinkForPhoto(newPhoto, fmt.Sprintf("https://www.twitch.tv/%s", user.Login), "Open the channel", updateInfo.Message.MessageID)

	return newPhoto
}
