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
	somethingWrong = "Oops, something wrong, try again later or contact my creator"
	invalidReq     = "Invalid request, try again. "

	// Time constants
	messageTimeoutDuration = 15 * time.Second
	updateTimeoutSeconds   = 60

	// Background sync process name
	telegramUpdatesCheckBGSync = "telegramUpdatesCheck_BGSync"
)

type teleCommands string

const (
	startCommand                     teleCommands = "/start"
	pingCommand                      teleCommands = "/ping"
	commands                         teleCommands = "/commands"
	twitchUserCommand                teleCommands = "/user_info"
	twitchStreamNotifi               teleCommands = "/stream_notify"
	twitchCancelStreamNotifi         teleCommands = "/cancel_stream_notify"
	twitchFollowedStreamNotify       teleCommands = "/followed_notify"
	twitchCancelFollowedStreamNotify teleCommands = "/cancel_followed_notify"
)

// commandHandler represents a function that handles a specific command
type commandHandler func(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error)

type TelegramUpdatesCheckService struct {
	twitchClient *twitch_client.TwitchClient
	fClient      *fileClient.FileClient

	dbRepo                *dbRepository.DBRepository
	notificationService   *notificationService.TwitchNotificationService
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService

	telegramService *telegram_service.TelegramService

	twitchOauthClient *twitch_oauth_client.TwitchOauthClient

	// Command router for efficient command handling
	commandHandlers map[teleCommands]commandHandler
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
	service := &TelegramUpdatesCheckService{
		twitchClient:          twitchClient,
		fClient:               fClient,
		dbRepo:                dbRepo,
		notificationService:   notifiService,
		twitchUserAuthservice: twitchUserAuthservice,
		telegramService:       telegramService,
		twitchOauthClient:     twitchOauthClient,
	}

	// Initialize command handlers
	service.initializeCommandHandlers()

	return service, nil
}

// initializeCommandHandlers sets up the command routing map
func (tmcs *TelegramUpdatesCheckService) initializeCommandHandlers() {
	tmcs.commandHandlers = map[teleCommands]commandHandler{
		startCommand:                     tmcs.handleStart,
		pingCommand:                      tmcs.handlePing,
		commands:                         tmcs.handleCommands,
		twitchUserCommand:                tmcs.handleTwitchUser,
		twitchStreamNotifi:               tmcs.handleTwitchStreamNotification,
		twitchCancelStreamNotifi:         tmcs.handleTwitchCancelStreamNotification,
		twitchFollowedStreamNotify:       tmcs.handleTwitchFollowedNotification,
		twitchCancelFollowedStreamNotify: tmcs.handleTwitchCancelFollowedNotification,
	}
}

func (tmcs *TelegramUpdatesCheckService) Sync(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return fmt.Errorf("failed to create bot API: %w", err)
	}

	logrus.Infof("Authorized on account %s", bot.Self.UserName)

	reader := tgbotapi.NewUpdate(0)
	reader.Timeout = updateTimeoutSeconds

	updates := bot.GetUpdatesChan(reader)

	for updateInfo := range updates {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := tmcs.processUpdate(ctx, updateInfo, bot); err != nil {
				logrus.Errorf("failed to process update: %v", err)
			}
		}
	}

	return nil
}

// processUpdate handles a single update from Telegram
func (tmcs *TelegramUpdatesCheckService) processUpdate(
	ctx context.Context,
	updateInfo tgbotapi.Update,
	bot *tgbotapi.BotAPI,
) error {
	if updateInfo.Message == nil {
		return nil
	}

	message := updateInfo.Message
	logrus.Infof("[%s] %s", message.From.UserName, message.Text)

	// Check if message is too old
	if tmcs.isMessageTooOld(message.Date) {
		return tmcs.handleOldMessage(ctx, message, bot)
	}

	// Process command
	return tmcs.handleCommand(ctx, updateInfo, bot)
}

// isMessageTooOld checks if a message is older than the allowed timeout
func (tmcs *TelegramUpdatesCheckService) isMessageTooOld(messageDate int) bool {
	messageTime := time.Unix(int64(messageDate), 0)
	return messageTime.Add(messageTimeoutDuration).Before(time.Now())
}

// handleOldMessage responds to old messages
func (tmcs *TelegramUpdatesCheckService) handleOldMessage(ctx context.Context, message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I took a little nap ☺️ . Now I'm awake and ready to go! 😎")
	msg.ReplyToMessageID = message.MessageID

	if err := sendMsgToTelegram(ctx, msg, bot); err != nil {
		return fmt.Errorf("failed to send old message response: %w", err)
	}

	logrus.Infof("Skipped old message from user %s, message time %s, current time %s",
		message.From.UserName,
		time.Unix(int64(message.Date), 0),
		time.Now())

	return nil
}

// handleCommand routes and processes commands
func (tmcs *TelegramUpdatesCheckService) handleCommand(ctx context.Context, updateInfo tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	text := updateInfo.Message.Text

	// Find matching command
	for command, handler := range tmcs.commandHandlers {
		if strings.HasPrefix(text, string(command)) {
			return tmcs.executeCommand(ctx, updateInfo, bot, handler)
		}
	}

	// No command matched - could log or handle as needed
	logrus.Debugf("No command handler found for text: %s", text)
	return nil
}

// executeCommand executes a command handler and sends the response
func (tmcs *TelegramUpdatesCheckService) executeCommand(ctx context.Context, updateInfo tgbotapi.Update, bot *tgbotapi.BotAPI, handler commandHandler) error {
	result, err := handler(ctx, updateInfo)
	if err != nil {
		logrus.Errorf("Command execution error: %v", err)
		// Send error message to user
		msg := tgbotapi.NewMessage(updateInfo.Message.Chat.ID, somethingWrong)
		msg.ReplyToMessageID = updateInfo.Message.MessageID
		return sendMsgToTelegram(ctx, msg, bot)
	}

	// Send response based on result type
	switch response := result.(type) {
	case tgbotapi.MessageConfig:
		return sendMsgToTelegram(ctx, response, bot)
	case tgbotapi.PhotoConfig:
		return sendPhotoToTelegram(ctx, response, bot)
	default:
		logrus.Warnf("Unknown response type: %T", result)
		return nil
	}
}

// Command handlers
func (tmcs *TelegramUpdatesCheckService) handleStart(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	return tmcs.start(ctx, updateInfo)
}

func (tmcs *TelegramUpdatesCheckService) handlePing(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	msg := tgbotapi.NewMessage(updateInfo.Message.Chat.ID, "pong")
	msg.ReplyToMessageID = updateInfo.Message.MessageID
	return msg, nil
}

func (tmcs *TelegramUpdatesCheckService) handleCommands(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	return tmcs.commands(ctx, updateInfo)
}

func (tmcs *TelegramUpdatesCheckService) handleTwitchUser(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	photo, isFound := tmcs.twitchUserCase(ctx, updateInfo)
	if !isFound {
		msg := tgbotapi.NewMessage(photo.ChatID, photo.Caption)
		msg.ReplyToMessageID = photo.ReplyToMessageID
		return msg, nil
	}
	return photo, nil
}

func (tmcs *TelegramUpdatesCheckService) handleTwitchStreamNotification(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	return tmcs.twitchAddStreamNotification(ctx, updateInfo)
}

func (tmcs *TelegramUpdatesCheckService) handleTwitchCancelStreamNotification(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	return tmcs.twitchCancelStreamNotification(ctx, updateInfo)
}

func (tmcs *TelegramUpdatesCheckService) handleTwitchFollowedNotification(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	return tmcs.twitchAddFollowedNotification(ctx, updateInfo)
}

func (tmcs *TelegramUpdatesCheckService) handleTwitchCancelFollowedNotification(ctx context.Context, updateInfo tgbotapi.Update) (interface{}, error) {
	return tmcs.twitchCancelFollowedNotification(ctx, updateInfo)
}

func sendMsgToTelegram(_ context.Context, resp tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) error {
	_, err := bot.Send(resp)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	return nil
}

func sendPhotoToTelegram(_ context.Context, resp tgbotapi.PhotoConfig, bot *tgbotapi.BotAPI) error {
	_, err := bot.Send(resp)
	if err != nil {
		return fmt.Errorf("failed to send telegram photo: %w", err)
	}
	return nil
}

func (tmcs *TelegramUpdatesCheckService) SyncBg(ctx context.Context, syncInterval time.Duration) {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	logrus.Infof("Starting background sync process: %s", telegramUpdatesCheckBGSync)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("Stopping background sync process: %s", telegramUpdatesCheckBGSync)
			return
		case <-ticker.C:
			logrus.Debugf("Starting background sync iteration: %s", telegramUpdatesCheckBGSync)

			if err := tmcs.Sync(ctx); err != nil {
				logrus.Errorf("Background sync failed: %v", err)
				continue
			}

			logrus.Debugf("Background sync completed successfully: %s", telegramUpdatesCheckBGSync)
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
