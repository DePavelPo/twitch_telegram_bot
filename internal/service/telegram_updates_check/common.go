package telegram_updates_check

import (
	"context"
	"fmt"
	"strings"
	"twitch_telegram_bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	greetingMessage   = "Greetings! The bot provides the functionality for interacting with Twitch streaming platform"
	commandListHeader = "Bot's command list:"
	commandFormat     = "%s - %s"
)

// start handles the /start command
func (tmcs *TelegramUpdatesCheckService) start(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.MessageConfig{}
	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	// Build the response message
	response := tmcs.buildCommandListResponse(ctx, greetingMessage, true)
	msg.Text = response

	return msg, nil
}

// commands handles the /commands command
func (tmcs *TelegramUpdatesCheckService) commands(
	ctx context.Context,
	updateInfo tgbotapi.Update,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.MessageConfig{}
	msg.ChatID = updateInfo.Message.Chat.ID
	msg.ReplyToMessageID = updateInfo.Message.MessageID

	// Build the response message
	response := tmcs.buildCommandListResponse(ctx, "", false)
	msg.Text = response

	return msg, nil
}

// buildCommandListResponse builds the command list response message
func (tmcs *TelegramUpdatesCheckService) buildCommandListResponse(
	ctx context.Context,
	prefix string,
	includePrefix bool,
) string {
	var builder strings.Builder

	// Add prefix if provided and requested
	if includePrefix && prefix != "" {
		builder.WriteString(prefix)
		builder.WriteString("\n")
	}

	// Get bot commands
	teleCommands, err := tmcs.getBotCommands(ctx)
	if err != nil {
		logrus.Errorf("Failed to get bot commands: %v", err)
		return somethingWrong
	}

	// If no commands available, return appropriate message
	if teleCommands == nil || len(teleCommands.Commands) == 0 {
		if includePrefix {
			return builder.String() + "No commands available at the moment."
		}
		return "No commands available at the moment."
	}

	// Add command list header
	builder.WriteString(commandListHeader)

	// Add each command
	for _, teleCommand := range teleCommands.Commands {
		builder.WriteString("\n")
		builder.WriteString(fmt.Sprintf(commandFormat, teleCommand.Command, teleCommand.Description))
	}

	return builder.String()
}

// getBotCommands safely retrieves bot commands with error handling
func (tmcs *TelegramUpdatesCheckService) getBotCommands(ctx context.Context) (*models.TeleBotCommands, error) {
	teleCommands, err := tmcs.telegramService.GetBotCommands(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bot commands")
	}
	return teleCommands, nil
}
