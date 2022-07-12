package telegram_updates_check

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"twitch_telegram_bot/internal/models"

	twitch_client "twitch_telegram_bot/internal/client/twitch-client"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	telegramUpdatesCheckBGSync = "telegramUpdatesCheck_BGSync"
	pingCommand                = "/ping"
	jokeCommand                = "/anec"
	twitchUserCommand          = "/twitch_user"
)

type TelegramUpdatesCheckService struct {
	twitchClient *twitch_client.TwitchClient
}

func NewTelegramUpdatesCheckService(twitchClient *twitch_client.TwitchClient) (*TelegramUpdatesCheckService, error) {
	return &TelegramUpdatesCheckService{
		twitchClient: twitchClient,
	}, nil
}

func (tmcs *TelegramUpdatesCheckService) Sync(ctx context.Context) error {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return err
	}

	// для подробных логов
	bot.Debug = true

	logrus.Printf("Authorized on account %s", bot.Self.UserName)

	reader := tgbotapi.NewUpdate(0)
	reader.Timeout = 60

	updates := bot.GetUpdatesChan(reader)

	for updateInfo := range updates {
		if updateInfo.Message != nil {
			logrus.Printf("[%s] %s", updateInfo.Message.From.UserName, updateInfo.Message.Text)

			timeAndZone := time.Unix(int64(updateInfo.Message.Date), 0)

			msg := tgbotapi.NewMessage(updateInfo.Message.Chat.ID, "")

			timeNow := time.Now()
			// TODO: подумать, как избежать дубликации ответа
			if timeAndZone.Add(time.Second * 12).Before(timeNow) {

				msg.Text = "Прошу прощения, я немного вздремнул ☺️ . Теперь я пробудился и готов к работе! 😎 "
				msg.ReplyToMessageID = updateInfo.Message.MessageID

				bot.Send(msg)

				logrus.Printf("skip reason: old time. User %s, message time %s, time now %s", updateInfo.Message.From.UserName, timeAndZone, timeNow)
				continue
			}

			// TODO: добавить валидацию

			// TODO: расширять функционал
			switch {
			case strings.HasPrefix(updateInfo.Message.Text, pingCommand):
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, jokeCommand):
				rand.Seed(time.Now().UnixNano())
				msg.Text = models.JokeList[rand.Intn(len(models.JokeList))]

			// TODO: унести в отдельную функцию
			case strings.HasPrefix(updateInfo.Message.Text, twitchUserCommand):
				userLogin := updateInfo.Message.Text[len(fmt.Sprintf("%s ", twitchUserCommand)):]
				users, err := tmcs.twitchClient.GetUserInfo(ctx, os.Getenv("TWITCH_BEARER"), []string{userLogin})
				if err != nil {
					msg.Text = "Ой, что-то пошло не так, повторите попытку позже или обратитесь к моему автору"
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
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
					userType = "пользователь"
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
					userBroadcasterType = "пользователь"
				}

				// TODO: подкорректировать отображение
				msg.Text = fmt.Sprintf(`Пользователь: %s
				Дата создания аккаунта: %s
				Тип пользователя: %s
				Тип стримера: %s
				%s
				`,
					user.DisplayName,
					accCreatedTime.Format("2006.02.01 15:04:05"),
					userType,
					userBroadcasterType,
					fmt.Sprintf("https://www.twitch.tv/%s", user.Login))

				msg.ReplyToMessageID = updateInfo.Message.MessageID
			}

			bot.Send(msg)
		}
	}

	return nil
}

func (tmcs *TelegramUpdatesCheckService) SyncBg(ctx context.Context, syncInterval time.Duration) {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("stoping bg %s process", telegramUpdatesCheckBGSync)
			return
		case <-ticker.C:
			logrus.Infof("started bg %s process", telegramUpdatesCheckBGSync)
			err := tmcs.Sync(ctx)
			if err != nil {
				logrus.Info("could not check telegram updates")
				continue
			}
			logrus.Info("telegram updates check was complited")
		}
	}

}
