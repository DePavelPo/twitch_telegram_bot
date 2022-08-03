package telegram_updates_check

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var exampleText string = `Пример запроса:
	/twitch_user welovegames
	/twitch_user WELOVEGAMES
	`

func (tmcs *TelegramUpdatesCheckService) TwitchUserCase(ctx context.Context, msg tgbotapi.MessageConfig, updateInfo tgbotapi.Update) tgbotapi.MessageConfig {

	commandText := updateInfo.Message.Text[len(fmt.Sprintf("%s", twitchUserCommand)):]

	userLogin, isValid := validateText(commandText)
	if userLogin == nil || !isValid {
		msg.Text = `Не корректно составленный запрос, повторите попытку. ` + exampleText
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return msg
	}

	users, err := tmcs.twitchClient.GetUserInfo(ctx, []string{*userLogin})
	if err != nil {
		logrus.Error(err)
		msg.Text = "Ой, что-то пошло не так, повторите попытку позже или обратитесь к моему автору"
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return msg
	}
	// 380772110
	if users == nil {
		msg.Text = "Пользователь не найден"
		msg.ReplyToMessageID = updateInfo.Message.MessageID
	}
	if len(users.Data) < 1 {
		msg.Text = "Пользователь не найден"
		msg.ReplyToMessageID = updateInfo.Message.MessageID
	}

	user := users.Data[0]
	accCreatedTime := user.CreatedAt

	// отображаем по МСК
	location := time.FixedZone("MSK", 3*60*60)
	accCreatedTime = accCreatedTime.In(location)

	var userType string
	switch user.Type {
	case "staff":
		userType = "сотрудник твича"
		break
	case "admin":
		userType = "админ твича"
		break
	case "global_mod":
		userType = "глобальный администратор"
		break
	default:
		userType = "простой смертный"
	}

	var userBroadcasterType string
	switch user.BroadcasterType {
	case "partner":
		userBroadcasterType = "партнер"
		break
	case "affiliate":
		userBroadcasterType = "компаньон"
		break
	default:
		userBroadcasterType = "простой смертный"
	}

	msg.Text = fmt.Sprintf(`
	Информация о пользователе:
	Пользователь: %s
	Дата создания аккаунта: %s
	Тип пользователя: %s
	Тип стримера: %s
	`,
		user.DisplayName,
		accCreatedTime.Format("2006.01.02 15:04:05"),
		userType,
		userBroadcasterType)

	msg.ReplyToMessageID = updateInfo.Message.MessageID

	var streamStatus = "не определенно"
	streams, err := tmcs.twitchClient.GetActiveStreamInfoByUsers(ctx, []string{*userLogin})
	if err == nil || streams != nil {
		if len(streams.StreamInfo) < 1 {
			streamStatus = "оффлайн"
		} else {
			stream := streams.StreamInfo[0]

			if stream.UserId == *userLogin || stream.UserLogin == *userLogin || stream.UserName == *userLogin {
				streamStatus = "онлайн"
			}
		}
	}

	msg.Text = fmt.Sprintf(`
	%s
	Статус стрима: %s
	%s
	`,
		msg.Text,
		streamStatus,
		fmt.Sprintf("https://www.twitch.tv/%s", user.Login))

	return msg
}
