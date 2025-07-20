package telegram_updates_check

import (
	"context"
	"fmt"
	"time"

	"twitch_telegram_bot/internal/models"
	formater "twitch_telegram_bot/internal/utils/formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var userCustomExampleText string = `Request example:
	%s welovegames
	%s WELOVEGAMES
	`

func (tmcs *TelegramUpdatesCheckService) twitchUserCase(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (photo tgbotapi.PhotoConfig, isFound bool) {
	photo.ChatID = updateInfo.Message.Chat.ID
	photo.ReplyToMessageID = updateInfo.Message.MessageID

	commandText := updateInfo.Message.Text[len(fmt.Sprint(twitchUserCommand)):]

	userLogin, isValid := validateText(commandText)
	if !isValid {
		photo.Caption = invalidReq + fmt.Sprintf(userCustomExampleText, twitchUserCommand, twitchUserCommand)
		return
	}

	userLogin = formater.ToLower(userLogin)

	users, err := tmcs.twitchClient.GetUserInfo(ctx, []string{userLogin})
	if err != nil {
		logrus.Errorf("get user info failed: %s", err.Error())
		photo.Caption = somethingWrong
		return
	}

	if (users == nil) || (len(users.Data) < 1) {
		photo.Caption = "User not found"
		return
	}

	// we sure we've got user info
	// and we can work with it
	isFound = true

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
	Account created: %s (GMT+3)
	User type: %s
	Streamer type: %s
	`,
		user.DisplayName,
		accCreatedTime.Format("2006.01.02 15:04:05"),
		userType,
		userBroadcasterType)

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

	newPhoto := tgbotapi.NewPhoto(updateInfo.Message.Chat.ID, tgbotapi.FileURL(user.ProfileImageUrl))
	newPhoto.ReplyToMessageID = updateInfo.Message.MessageID
	newPhoto.Caption = photo.Caption

	newPhoto = formater.CreateTelegramSingleButtonLinkForPhoto(newPhoto,
		fmt.Sprintf("%s/%s", models.TwitchWWWSchemeHost, user.Login), "Open the channel", updateInfo.Message.MessageID)

	return newPhoto, isFound
}
