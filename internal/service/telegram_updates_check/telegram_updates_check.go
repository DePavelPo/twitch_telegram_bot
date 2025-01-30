package telegram_updates_check

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	notificationService "twitch_telegram_bot/internal/service/notification"
	twitchUserAuthservice "twitch_telegram_bot/internal/service/twitch-user-authorization"

	fileClient "twitch_telegram_bot/internal/client/file"
	twitch_client "twitch_telegram_bot/internal/client/twitch-client"
	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"

	telegram_service "twitch_telegram_bot/internal/service/telegram"

	dbRepository "twitch_telegram_bot/db/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	somethingWrong string = "Oops, something wrong, try again later or contact my creator"
	invalidReq     string = "Invalid request, try again. "
)

type teleCommands string

const (
	telegramUpdatesCheckBGSync                    = "telegramUpdatesCheck_BGSync"
	startCommand                     teleCommands = "/start"
	pingCommand                      teleCommands = "/ping"
	commands                         teleCommands = "/commands"
	twitchUserCommand                teleCommands = "/user_info"
	twitchStreamNotifi               teleCommands = "/stream_notify"
	twitchCancelStreamNotifi         teleCommands = "/cancel_stream_notify"
	twitchFollowedStreamNotify       teleCommands = "/followed_notify"
	twitchCancelFollowedStreamNotify teleCommands = "/cancel_followed_notify"
)

type TelegramUpdatesCheckService struct {
	twitchClient *twitch_client.TwitchClient
	fClient      *fileClient.FileClient

	dbRepo                *dbRepository.DBRepository
	notificationService   *notificationService.TwitchNotificationService
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService

	telegramService *telegram_service.TelegramService

	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

func NewTelegramUpdatesCheckService(
	twitchClient *twitch_client.TwitchClient,
	fClient *fileClient.FileClient,
	dbRepo *dbRepository.DBRepository,
	notifiService *notificationService.TwitchNotificationService,
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService,
	telegramService *telegram_service.TelegramService,
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient,
) (*TelegramUpdatesCheckService, error) {
	return &TelegramUpdatesCheckService{
		twitchClient:          twitchClient,
		fClient:               fClient,
		dbRepo:                dbRepo,
		notificationService:   notifiService,
		twitchUserAuthservice: twitchUserAuthservice,
		telegramService:       telegramService,
		twitchOauthClient:     twitchOauthClient,
	}, nil
}

func (tmcs *TelegramUpdatesCheckService) Sync(ctx context.Context) error {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return err
	}

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

				msg.Text = "Sorry, I took a little nap ☺️ . Now I'm awake and ready to go! 😎 "
				msg.ReplyToMessageID = updateInfo.Message.MessageID

				sendMsgToTelegram(ctx, msg, bot)

				logrus.Printf("skip reason: old time. User %s, message time %s, time now %s", updateInfo.Message.From.UserName, timeAndZone, timeNow)
				continue
			}

			// TODO: добавить валидацию
			// TODO: расширять функционал

			switch {
			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(startCommand)):

				msg, err := tmcs.start(ctx, updateInfo)
				if err != nil {
					logrus.Errorf("start error: %s", err.Error())
				}

				// send result message to telegram bot
				sendMsgToTelegram(ctx, msg, bot)

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(pingCommand)):
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

				sendMsgToTelegram(ctx, msg, bot)

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(commands)):

				msg, err := tmcs.commands(ctx, updateInfo)
				if err != nil {
					logrus.Errorf("commands error: %s", err.Error())
				}

				// send result message to telegram bot
				sendMsgToTelegram(ctx, msg, bot)

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchUserCommand)):

				// get user info from twitch and prepare data
				if photo, isFound := tmcs.twitchUserCase(ctx, updateInfo); !isFound {
					msg.ChatID = photo.ChatID
					msg.ReplyToMessageID = photo.ReplyToMessageID
					msg.Text = photo.Caption

					// send info without a picture to telegram bot
					sendMsgToTelegram(ctx, msg, bot)

				} else {

					// send user info with a picture to telegram bot
					sendPhotoToTelegram(ctx, photo, bot)
				}

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchStreamNotifi)):

				// go to make a task that notice about channel live streams
				msg, err := tmcs.twitchAddStreamNotification(ctx, updateInfo)
				if err != nil {
					logrus.Errorf("twitch stream notification error: %s", err.Error())
				}

				// send result message to telegram bot
				sendMsgToTelegram(ctx, msg, bot)

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchCancelStreamNotifi)):

				// go to cancel a task that notice about channel live streams
				msg, err := tmcs.twitchCancelStreamNotification(ctx, updateInfo)
				if err != nil {
					logrus.Errorf("twitch cancel stream notification error: %s", err.Error())
				}

				// send result message to telegram bot
				sendMsgToTelegram(ctx, msg, bot)

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchFollowedStreamNotify)):

				// go to make a task that notice about channels' live streams from user following list on Twitch
				msg, err := tmcs.twitchAddFollowedNotification(ctx, updateInfo)
				if err != nil {
					logrus.Errorf("twitch followed stream notification error: %s", err.Error())
				}

				// send result message to telegram bot
				sendMsgToTelegram(ctx, msg, bot)

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchCancelFollowedStreamNotify)):

				// go to cancel a task that notice about channels' live streams from user following list on Twitch
				msg, err := tmcs.twitchCancelFollowedNotification(ctx, updateInfo)
				if err != nil {
					logrus.Errorf("twitch cancel followed stream notification error: %s", err.Error())
				}

				// send result message to telegram bot
				sendMsgToTelegram(ctx, msg, bot)

			default:

			}

		}

	}

	return nil
}

func sendMsgToTelegram(_ context.Context, resp tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
	_, err := bot.Send(resp)
	if err != nil {
		logrus.Errorf("telegram send message error: %s", err.Error())
	}
}

func sendPhotoToTelegram(_ context.Context, resp tgbotapi.PhotoConfig, bot *tgbotapi.BotAPI) {
	_, err := bot.Send(resp)
	if err != nil {
		logrus.Errorf("telegram send message error: %s", err.Error())
	}
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

func validateText(text string) (string, bool) {

	words := strings.Fields(text)

	if len(words) != 1 || words[0] == "" {
		return "", false
	}

	return words[0], true
}
