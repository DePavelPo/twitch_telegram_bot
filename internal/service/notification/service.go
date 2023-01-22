package notification

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	twitch_client "twitch_telegram_bot/internal/client/twitch-client"
	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"

	dbRepository "twitch_telegram_bot/db/repository"
)

type TwitchNotificationService struct {
	dbRepo            *dbRepository.DBRepository
	twitchClient      *twitch_client.TwitchClient
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

func NewTwitchNotificationService(
	dbRepo *dbRepository.DBRepository,
	twitchClient *twitch_client.TwitchClient,
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient,
) (*TwitchNotificationService, error) {
	service := &TwitchNotificationService{
		dbRepo:            dbRepo,
		twitchClient:      twitchClient,
		twitchOauthClient: twitchOauthClient,
	}

	ctx := context.Background()
	err := service.Sync(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Sync")
	}

	return service, nil
}

func (tn *TwitchNotificationService) SyncBg(ctx context.Context, syncInterval time.Duration) {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("stoping bg %s process", twitchNotificationBGSync)
			return
		case <-ticker.C:
			logrus.Infof("started bg %s process", twitchNotificationBGSync)
			err := tn.Sync(ctx)
			if err != nil {
				logrus.Infof("could not check twitch token: %v", err)
				continue
			}
			logrus.Info("twitch token check was complited")
		}
	}
}
