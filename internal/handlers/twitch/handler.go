package twitch_handler

import (
	twitch_service "twitch_telegram_bot/internal/service/twitch"
	twitchUserAuthservice "twitch_telegram_bot/internal/service/twitch-user-authorization"
)

type TwitchHandler struct {
	twitchService         *twitch_service.TwitchService
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService
}

func NewTwitchHandler(
	twitchService *twitch_service.TwitchService,
	twitchUserAuthservice *twitchUserAuthservice.TwitchUserAuthorizationService,
) *TwitchHandler {
	return &TwitchHandler{
		twitchService:         twitchService,
		twitchUserAuthservice: twitchUserAuthservice,
	}
}
