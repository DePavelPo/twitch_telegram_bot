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

	twitch_client "twitch_telegram_bot/internal/client/twitch-client"

	telegram_service "twitch_telegram_bot/internal/service/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type teleCommands string

const (
	telegramUpdatesCheckBGSync              = "telegramUpdatesCheck_BGSync"
	pingCommand                teleCommands = "/ping"
	commands                   teleCommands = "/commands"
	jokeCommand                teleCommands = "/anec"
	twitchUserCommand          teleCommands = "/twitch_user"
	twitchBanTest              teleCommands = "/twitch_ban_test"
	twitchStreamNotifi         teleCommands = "/twitch_stream_notification"
	twitchDropStreamNotifi     teleCommands = "/twitch_drop_stream_notification"
)

type TelegramUpdatesCheckService struct {
	twitchClient        *twitch_client.TwitchClient
	notificationService *notificationService.TwitchNotificationService

	telegramService *telegram_service.TelegramService
}

func NewTelegramUpdatesCheckService(
	twitchClient *twitch_client.TwitchClient,
	notifiService *notificationService.TwitchNotificationService,
	telegramService *telegram_service.TelegramService,
) (*TelegramUpdatesCheckService, error) {
	return &TelegramUpdatesCheckService{
		twitchClient:        twitchClient,
		notificationService: notifiService,
		telegramService:     telegramService,
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
			// TODO: добавить обработку комманды /start
			switch {
			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(pingCommand)):
				msg.Text = "pong"
				msg.ReplyToMessageID = updateInfo.Message.MessageID
				break

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(commands)):
				teleCommands, err := tmcs.telegramService.GetBotCommands(ctx)
				if err != nil {
					logrus.Infof("GetBotCommands error: %v", err)
					msg.Text = "Ой, что-то пошло не так, повторите попытку позже или обратитесь к моему автору"
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = `Список комманд бота:`

				if teleCommands != nil {
					for _, teleCommand := range teleCommands.Commands {
						msg.Text = fmt.Sprintf(
							`
							%s
							%s - %s
							`, msg.Text, teleCommand.Command, teleCommand.Description,
						)
					}

				}

				msg.ReplyToMessageID = updateInfo.Message.MessageID
				break

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(jokeCommand)):

				msg.Text = fmt.Sprintf(`
				Внимание, анекдот!
				
				%s`,
					models.JokeList[rand.Intn(len(models.JokeList))])
				break

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchBanTest)):
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
				break

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchUserCommand)):

				msg = tmcs.TwitchUserCase(ctx, msg, updateInfo)
				break

			// TODO: кастомизировать exampleText

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchStreamNotifi)):

				chatId := updateInfo.Message.Chat.ID

				commandText := updateInfo.Message.Text[len(fmt.Sprintf("%s", twitchStreamNotifi)):]

				userLogin, isValid := validateText(commandText)
				if userLogin == nil || !isValid {
					msg.Text = `Не корректно составленный запрос, повторите попытку. ` + exampleText
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				err := tmcs.notificationService.AddTwitchNotification(ctx, uint64(chatId), *userLogin)
				if err != nil {
					logrus.Infof("Add twitch notification request error: %v", err)
					msg.Text = "Ой, что-то пошло не так, повторите попытку позже или обратитесь к моему автору"
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = "Запрос успешно принят! Теперь в этот канал будут приходить уведомления о трансляции на указанном вами twitch канале"
				msg.ReplyToMessageID = updateInfo.Message.MessageID
				break

			case strings.HasPrefix(updateInfo.Message.Text, fmt.Sprint(twitchDropStreamNotifi)):

				chatId := updateInfo.Message.Chat.ID

				commandText := updateInfo.Message.Text[len(fmt.Sprintf("%s", twitchDropStreamNotifi)):]

				userLogin, isValid := validateText(commandText)
				if userLogin == nil || !isValid {
					msg.Text = `Не корректно составленный запрос, повторите попытку. ` + exampleText
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				err := tmcs.notificationService.SetInactiveNotification(ctx, uint64(chatId), *userLogin)
				if err != nil {
					if err.Error() == "notification not found" {
						logrus.Infof("notification by chatId %d user %s not found", chatId, *userLogin)
						msg.Text = "Заявок на уведомления по этому каналу не найдено. Возможно неправильно указано наименование или такая заявка не создавалась"
						msg.ReplyToMessageID = updateInfo.Message.MessageID
						break
					}
					logrus.Infof("Set inactive twitch notification error: %v", err)
					msg.Text = "Ой, что-то пошло не так, повторите попытку позже или обратитесь к моему автору"
					msg.ReplyToMessageID = updateInfo.Message.MessageID
					break
				}

				msg.Text = "Уведомления по указанному twitch каналу успешно отключены"
				msg.ReplyToMessageID = updateInfo.Message.MessageID
				break
			}

			_, err = bot.Send(msg)
			if err != nil {
				logrus.Infof("/twitch_user: telegram send message error: %v", err)
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

func validateText(text string) (str *string, isValid bool) {

	words := strings.Fields(text)

	if len(words) != 1 {
		return nil, false
	}

	return &words[0], true
}
