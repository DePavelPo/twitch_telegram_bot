package telegram_updates_check

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"twitch_telegram_bot/internal/models"

	notificationService "twitch_telegram_bot/internal/service/notification"
	twitchUserAuthservice "twitch_telegram_bot/internal/service/twitch-user-authorization"

	twitch_client "twitch_telegram_bot/internal/client/twitch-client"
	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"

	telegram_service "twitch_telegram_bot/internal/service/telegram"

	text_formater "twitch_telegram_bot/internal/utils/text-formater"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	somethingWrong string = "Oops, something wrong, try again later or contact my creator"
	invalidReq     string = "Invalid request, try again. "
)

type teleCommands string

const (
	telegramUpdatesCheckBGSync                  = "telegramUpdatesCheck_BGSync"
	startCommand                   teleCommands = "/start"
	pingCommand                    teleCommands = "/ping"
	commands                       teleCommands = "/commands"
	jokeCommand                    teleCommands = "/anec"
	twitchUserCommand              teleCommands = "/twitch_user"
	twitchBanTest                  teleCommands = "/twitch_ban_test"
	twitchStreamNotifi             teleCommands = "/twitch_stream_notify"
	twitchDropStreamNotifi         teleCommands = "/twitch_drop_stream_notify"
	twitchFollowedStreamNotify     teleCommands = "/twitch_followed_notify"
	twitchDropFollowedStreamNotify teleCommands = "/twitch_drop_followed_notify"
)

type TelegramUpdatesCheckService struct {
	twitchClient          *twitch_client.TwitchClient
	notificationService   *notificationService.TwitchNotificationService
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService

	telegramService *telegram_service.TelegramService

	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

func NewTelegramUpdatesCheckService(
	twitchClient *twitch_client.TwitchClient,
	notifiService *notificationService.TwitchNotificationService,
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService,
	telegramService *telegram_service.TelegramService,
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient,
) (*TelegramUpdatesCheckService, error) {
	return &TelegramUpdatesCheckService{
		twitchClient:          twitchClient,
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

				msg.Text = "I'm sorry I took a little nap ☺️ . Now I'm awake and ready to go! 😎 "
				msg.ReplyToMessageID = updateInfo.Message.MessageID

				bot.Send(msg)

				logrus.Printf("skip reason: old time. User %s, message time %s, time now %s", updateInfo.Message.From.UserName, timeAndZone, timeNow)
				continue
			}

			// TODO: добавить валидацию
			rand.Seed(time.Now().UnixNano())
			// TODO: расширять функционал
			// TODO: раскидать все кейсы по отдельным функциям

			chatId := updateInfo.Message.Chat.ID

			switch {
			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(startCommand)):

				msg.Text = `Greetings! The bot provides functionality for interacting with Twitch streaming platform
				`

				teleCommands, err := tmcs.telegramService.GetBotCommands(ctx)
				if err != nil {
					logrus.Infof("GetBotCommands error: %v", err)
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = fmt.Sprintf(
					`%s
					%s
					`, msg.Text, "Bot's command list:")

				if teleCommands != nil {
					for _, teleCommand := range teleCommands.Commands {
						msg.Text = fmt.Sprintf(
							`
							%s
							%s - %s`, msg.Text, teleCommand.Command, teleCommand.Description,
						)
					}

				}

				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(pingCommand)):
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(commands)):
				teleCommands, err := tmcs.telegramService.GetBotCommands(ctx)
				if err != nil {
					logrus.Infof("GetBotCommands error: %v", err)
					msg.Text = somethingWrong
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = `Bot's command list:
				
				`

				if teleCommands != nil {
					for _, teleCommand := range teleCommands.Commands {
						msg.Text = fmt.Sprintf(
							`
							%s
							%s - %s`, msg.Text, teleCommand.Command, teleCommand.Description,
						)
					}

				}

				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(jokeCommand)):

				msg.Text = fmt.Sprintf(`
				Attention! joke!
				
				%s`,
					models.JokeList[rand.Intn(len(models.JokeList))])

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchBanTest)):
				var emote string

				chance := rand.Intn(101)
				switch {
				case chance == 0:
					emote = "😩"
				case chance > 0 && chance <= 25:
					emote = "🤔"
				case chance > 25 && chance <= 50:
					emote = "😮"
				case chance > 50 && chance <= 75:
					emote = "😃"
				case chance > 75 && chance <= 99:
					emote = "🤯"
				default:
					emote = "😎"
				}

				msg.Text = fmt.Sprintf("Your chance to get banned on Twitch = %d%% %s", chance, emote)

				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchUserCommand)):

				msg = tmcs.TwitchUserCase(ctx, msg, updateInfo)

			// TODO: кастомизировать exampleText

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchStreamNotifi)):

				chatId := updateInfo.Message.Chat.ID

				commandText := updateInfo.Message.Text[len(fmt.Sprint(twitchStreamNotifi)):]

				userLogin, isValid := validateText(commandText)
				if userLogin == "" || !isValid {
					msg.Text = invalidReq + exampleText
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				userLogin = text_formater.ToLower(userLogin)

				err := tmcs.notificationService.AddTwitchNotification(ctx, uint64(chatId), userLogin, models.NotificationByUser)
				if err != nil {
					logrus.Infof("Add twitch notification request error: %v", err)
					msg.Text = somethingWrong
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = "Request successfully accepted! This channel will now receive stream notifications from Twitch channel that you specified"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchDropStreamNotifi)):

				commandText := updateInfo.Message.Text[len(fmt.Sprint(twitchDropStreamNotifi)):]

				userLogin, isValid := validateText(commandText)
				if userLogin == "" || !isValid {
					msg.Text = invalidReq + exampleText
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				userLogin = text_formater.ToLower(userLogin)

				err := tmcs.notificationService.SetInactiveNotificationByUser(ctx, uint64(chatId), userLogin)
				if err != nil {
					if err.Error() == "notification not found" {
						logrus.Infof("notification by chatId %d user %s not found", chatId, userLogin)
						msg.Text = "No requests for notifications were found for this channel. Perhaps the name is incorrectly indicated or such request was not created"
						msg.ReplyToMessageID = updateInfo.Message.MessageID
						break
					}
					logrus.Infof("Set inactive twitch notification error: %v", err)
					msg.Text = somethingWrong
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = "Notifications were disabled successfully"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchFollowedStreamNotify)):

				data, err := tmcs.twitchUserAuthservice.CheckUserTokensByChat(ctx, uint64(chatId))
				if err != nil {
					logrus.Infof("CheckUserTokensByChat error: %v", err)
					msg.Text = somethingWrong
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				if data.Link != "" {

					logrus.Info(data.Link)

					msg = createTwitchOath2LinkResp(msg, data.Link, updateInfo.Message.MessageID)

					logrus.Info(msg)

					break
				}

				err = tmcs.notificationService.AddTwitchNotification(ctx, uint64(chatId), data.UserID, models.NotificationFollowed)
				if err != nil {
					logrus.Infof("Add twitch notification request error: %v", err)
					msg.Text = somethingWrong
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = "Request successfully accepted! This channel will now receive stream notifications from channels that you following"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchDropFollowedStreamNotify)):

				err := tmcs.notificationService.SetInactiveNotificationFollowed(ctx, uint64(chatId))
				if err != nil {
					if err.Error() == "notification not found" {
						logrus.Infof("followed streams notification by chatId %d not found", chatId)
						msg.Text = "No requests for notifications were found 😕"
						msg.ReplyToMessageID = updateInfo.Message.MessageID
						break
					}
					logrus.Infof("Set inactive twitch notification error: %v", err)
					msg.Text = somethingWrong
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = "Notifications were disabled successfully"
				msg.ReplyToMessageID = updateInfo.Message.MessageID

			}

			_, err = bot.Send(msg)
			if err != nil {
				logrus.Infof("telegram send message error: %v", err)
			}
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

func validateText(text string) (str string, isValid bool) {

	words := strings.Fields(text)

	if len(words) != 1 {
		return "", false
	}

	return words[0], true
}

func createTwitchOath2LinkResp(msg tgbotapi.MessageConfig, link string, messageID int) tgbotapi.MessageConfig {

	resp := "To use this functionality, follow the link and provide access to the necessary information"
	msg.Text = resp

	board := make([][]tgbotapi.InlineKeyboardButton, 1)
	for i := range board {
		board[i] = make([]tgbotapi.InlineKeyboardButton, 1)
	}

	board[0][0].Text = "Open Link"
	board[0][0].URL = &link

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: board,
	}

	msg.ReplyToMessageID = messageID

	return msg

}
