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
	twitchBanTest              = "/twitch_ban_test"
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
			if timeAndZone.Add(time.Second * 15).Before(timeNow) {

				msg.Text = "Прошу прощения, я немного вздремнул ☺️ . Теперь я пробудился и готов к работе! 😎 "
				msg.ReplyToMessageID = updateInfo.Message.MessageID

				bot.Send(msg)

				logrus.Printf("skip reason: old time. User %s, message time %s, time now %s", updateInfo.Message.From.UserName, timeAndZone, timeNow)
				continue
			}

			// TODO: добавить валидацию

			rand.Seed(time.Now().UnixNano())
			// TODO: расширять функционал
			// TODO: рандомайз на шанс быть забаненным на твиче
			switch {
			case strings.HasPrefix(updateInfo.Message.Text, pingCommand):
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID
				break

			case strings.HasPrefix(updateInfo.Message.Text, jokeCommand):

				msg.Text = fmt.Sprintf(`
				Внимание, анекдот!
				
				%s`,
					models.JokeList[rand.Intn(len(models.JokeList))])
				break

			case strings.HasPrefix(updateInfo.Message.Text, twitchBanTest):
				var emote string

				chance := rand.Intn(101)
				switch {
				case chance == 0:
					emote = "😩"
					break
				case chance > 0 && chance <= 25:
					emote = "🤔"
					break
				case chance > 25 && chance <= 50:
					emote = "😮"
					break
				case chance > 50 && chance <= 75:
					emote = "😃"
					break
				case chance > 75 && chance <= 99:
					emote = "🤯"
					break
				default:
					emote = "😎"
				}

				msg.Text = fmt.Sprintf("Твой шанс быть забанненым на твиче = %d%% %s", chance, emote)

				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, twitchUserCommand):
				msg = tmcs.TwitchUserCase(ctx, msg, updateInfo)
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

func validateText(text string) (str *string, isValid bool) {

	words := strings.Fields(text)

	if len(words) != 1 {
		return nil, false
	}

	return &words[0], true
}
