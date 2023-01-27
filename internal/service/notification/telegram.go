package notification

import (
	"context"
	"fmt"
	"os"
	"strings"
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

	var photoLink string
	// TODO: add default photo if current url not created
	if strings.Contains(stream.ThumbnailUrl, "{width}x{height}") {
		photoLink = strings.Replace(stream.ThumbnailUrl, "{width}x{height}", "1920x1080", -1)
	}

	photo := tgbotapi.NewPhoto(int64(chatId), tgbotapi.FileURL(photoLink))

	twitchLink := fmt.Sprintf("https://www.twitch.tv/%s", stream.UserLogin)

	photo.Caption = fmt.Sprintf(`
	▶️ %s stream is online!
	Title: %s,
	Current stream duration: %s,
	Count of viewers: %d
	`,
		stream.UserName,
		stream.Title,
		formater.CreateStreamDuration(stream.StartedAt),
		stream.ViewerCount)

	photo = formater.CreateTelegramSingleButtonLinkForPhoto(photo, twitchLink, "Open the channel", 0)

	_, err = bot.Send(photo)
	if err != nil {
		logrus.Infof("ThrowNotification: telegram send message error: %v", err)
	}

	return nil
}
