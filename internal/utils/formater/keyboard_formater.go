package formater

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateTelegramSingleButtonLink(msg tgbotapi.MessageConfig, link, text string, messageID int) tgbotapi.MessageConfig {

	board := make([][]tgbotapi.InlineKeyboardButton, 1)
	for i := range board {
		board[i] = make([]tgbotapi.InlineKeyboardButton, 1)
	}

	board[0][0].Text = text
	board[0][0].URL = &link

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: board,
	}

	if messageID != 0 {
		msg.ReplyToMessageID = messageID
	}

	return msg

}

func CreateTelegramSingleButtonLinkForPhoto(msg tgbotapi.PhotoConfig, link, text string, messageID int) tgbotapi.PhotoConfig {

	board := make([][]tgbotapi.InlineKeyboardButton, 1)
	for i := range board {
		board[i] = make([]tgbotapi.InlineKeyboardButton, 1)
	}

	board[0][0].Text = text
	board[0][0].URL = &link

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: board,
	}

	if messageID != 0 {
		msg.ReplyToMessageID = messageID
	}

	return msg

}
