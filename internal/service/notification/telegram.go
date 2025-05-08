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

	title := formater.TwitchTagToTelegram(stream.Title)
	photo.Caption = prepareCaption(stream.UserName, title, formater.CreateStreamDuration(stream.StartedAt), stream.ViewerCount)

	twitchLink := fmt.Sprintf("%s/%s", models.TwitchWWWSchemeHost, stream.UserLogin)
	photo = formater.CreateTelegramSingleButtonLinkForPhoto(photo, twitchLink, "Open the channel", 0)

	// using Markdown for hyperlinks
	photo.ParseMode = "Markdown"

	// trying to send message once
	_, err = bot.Send(photo)
	if err != nil {
		logrus.Infof("ThrowNotification: first try: telegram send message error: %v", err)

		title = formater.ClearTags(stream.Title)
		photo.Caption = prepareCaption(stream.UserName, title, formater.CreateStreamDuration(stream.StartedAt), stream.ViewerCount)

		// not using any parse mods
		photo.ParseMode = ""

		// trying to send message again but without tags
		_, err = bot.Send(photo)
		if err != nil {
			logrus.Infof("ThrowNotification: second try: telegram send message error: %v", err)
		}
	}

	return nil
}

func prepareCaption(userName, title, duration string, viewerCount uint64) string {
	return fmt.Sprintf(`
	▶️ %s stream is online!
	Title: %s,
	Current stream duration: %s,
	Count of viewers: %d
	`,
		userName,
		title,
		duration,
		viewerCount,
	)
}
