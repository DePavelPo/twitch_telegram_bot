package telegram_updates_check

import (
	"context"
	"math/rand"
	"os"
	"time"
	"twitch_telegram_bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	telegramUpdatesCheckBGSync = "telegramUpdatesCheck_BGSync"
	pingCommand                = "/ping"
	jokeCommand                = "/anec"
)

type TelegramUpdatesCheckService struct {
}

func NewTelegramUpdatesCheckService() (*TelegramUpdatesCheckService, error) {
	return &TelegramUpdatesCheckService{}, nil
}

func (tmcs *TelegramUpdatesCheckService) Sync(ctx context.Context) error {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return err
	}

	bot.Debug = true

	logrus.Printf("Authorized on account %s", bot.Self.UserName)

	reader := tgbotapi.NewUpdate(0)
	reader.Timeout = 60

	updates := bot.GetUpdatesChan(reader)

	// TODO: сделать так, чтобы при включении бота не реагировать на каждое скопившееся сообщение

	for updateInfo := range updates {
		if updateInfo.Message != nil {
			logrus.Printf("[%s] %s", updateInfo.Message.From.UserName, updateInfo.Message.Text)

			msg := tgbotapi.NewMessage(updateInfo.Message.Chat.ID, "")
			switch updateInfo.Message.Text {
			case pingCommand:
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case jokeCommand:
				rand.Seed(time.Now().UnixNano())
				msg.Text = models.JokeList[rand.Intn(len(models.JokeList))]
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
