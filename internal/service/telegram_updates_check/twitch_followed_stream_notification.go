package telegram_updates_check

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (tmcs *TelegramUpdatesCheckService) TwitchCreateOAuth2Link(ctx context.Context) string {

	url := "https://id.twitch.tv/oauth2/authorize?client_id=9ktj5w1ir11yf4s5t7guo1eku6i6vx&response_type=token&redirect_uri=http://localhost:3000&scope=user:read:follows&state=c3ab8aa609ea11e793ae92361f192671"

	logrus.Info(url)

	return url
}
