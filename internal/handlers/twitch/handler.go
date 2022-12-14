package twitch_handler

import twitch_service "twitch_telegram_bot/internal/service/twitch"

type TwitchHandler struct {
	twitchService *twitch_service.TwitchService
}

func NewTwitchHandler(twitchService *twitch_service.TwitchService) *TwitchHandler {
	return &TwitchHandler{
		twitchService: twitchService,
	}
}
