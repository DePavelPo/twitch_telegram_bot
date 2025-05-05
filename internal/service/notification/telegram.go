package notification

import (
	"context"
	"fmt"
	"os"

	// "strings"
	"twitch_telegram_bot/internal/models"

	formater "twitch_telegram_bot/internal/utils/formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func (tn *TwitchNotificationService) ThrowNotification(ctx context.Context, stream models.Stream, chatId uint64) error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return err
	}

	// TODO: find the way to use custom photos of channels
	photo := tgbotapi.NewPhoto(int64(chatId), tgbotapi.FilePath("./sources/twitch_image.png"))

	photo.Caption = fmt.Sprintf(`
	▶️ %s stream is online!
	Title: %s,
	Current stream duration: %s,
	Count of viewers: %d
	`,
		stream.UserName,
		formater.TwitchTagToTelegram(stream.Title),
		formater.CreateStreamDuration(stream.StartedAt),
		stream.ViewerCount)

	twitchLink := fmt.Sprintf("https://www.twitch.tv/%s", stream.UserLogin)
	photo = formater.CreateTelegramSingleButtonLinkForPhoto(photo, twitchLink, "Open the channel", 0)

	// using Markdown for hyperlinks
	photo.ParseMode = "Markdown"

	_, err = bot.Send(photo)
	if err != nil {
		logrus.Infof("ThrowNotification: telegram send message error: %v", err)
	}

	return nil
}
