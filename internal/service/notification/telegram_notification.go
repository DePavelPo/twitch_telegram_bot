package twitch_notification

import (
	"context"
	"fmt"
	"os"
	"twitch_telegram_bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func (tn *TwitchNotificationService) ThrowNotification(ctx context.Context, stream models.Stream, chatId uint64) error {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return err
	}

	// для подробных логов
	bot.Debug = true

	msg := tgbotapi.NewMessage(int64(chatId), "")

	twitchLink := fmt.Sprintf("https://www.twitch.tv/%s", stream.UserLogin)

	msg.Text = fmt.Sprintf(`
	Стрим пользователя %s онлайн!
	Заголовок: %s
	%s
	`,
		stream.UserName,
		stream.Title,
		twitchLink)

	_, err = bot.Send(msg)
	if err != nil {
		logrus.Infof("ThrowNotification: telegram send message error: %v", err)
	}

	return nil
}
