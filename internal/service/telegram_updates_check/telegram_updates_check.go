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

			// расширять функционал
			switch {
			case strings.HasPrefix(updateInfo.Message.Text, pingCommand):
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, jokeCommand):
				rand.Seed(time.Now().UnixNano())
				msg.Text = models.JokeList[rand.Intn(len(models.JokeList))]

			case strings.HasPrefix(updateInfo.Message.Text, twitchUserCommand):
				user := updateInfo.Message.Text[len(fmt.Sprintf("%s ", twitchUserCommand)):]
				userInfo, err := tmcs.twitchClient.GetUserInfo(ctx, os.Getenv("TWITCH_BEARER"), []string{user})
				if err != nil {
					msg.Text = "Ой, что-то пошло не так, повторите попытку позже или обратитесь к моему автору"
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				// TODO: поресерчить корректность перевода времени на корректную временную зону
				location := time.Now().Location()
				// TODO: сделать отдельную переменную под createdAt
				userInfo.Data[0].CreatedAt.In(location)

				// TODO: подкорректировать отображение
				msg.Text = fmt.Sprintf(
					`Пользователь: %s
				Дата создания аккаунта: %s
				%s
				`, userInfo.Data[0].DisplayName, userInfo.Data[0].CreatedAt.Format("2006.02.01 15:04:05"), fmt.Sprintf("https://www.twitch.tv/%s", userInfo.Data[0].Login))
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
