package notification

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	twitch_client "twitch_telegram_bot/internal/client/twitch-client"

)

type TwitchNotificationService struct {
	db           *sqlx.DB
	twitchClient *twitch_client.TwitchClient
}

func NewTwitchNotificationService(db *sqlx.DB, twitchClient *twitch_client.TwitchClient) (*TwitchNotificationService, error) {
	service := &TwitchNotificationService{
		db:           db,
		twitchClient: twitchClient,
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
