package twitch_token

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"

	dbRepository "twitch_telegram_bot/db/repository"
)

const (
	twitchTokeCheckBGSync = "twitchTokenCheck_BGSync"
	tokenInvalid          = "token invalid"
)

type TwitchTokenService struct {
	dbRepo            *dbRepository.DBRepository
	token             string
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

// TODO: прокидывать токен в другие модули
func NewTwitchTokenService(
	dbRepo *dbRepository.DBRepository,
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient,
) (*TwitchTokenService, error) {

	service := &TwitchTokenService{
		dbRepo:            dbRepo,
		twitchOauthClient: twitchOauthClient,
	}

	ctx := context.Background()
	err := service.Sync(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Sync")
	}

	return service, nil
}

func (tts *TwitchTokenService) GetCurrentToken(ctx context.Context) string {
	return tts.token
}

// TODO: подумать над оптимизацией
func (tts *TwitchTokenService) Sync(ctx context.Context) error {

	tx, err := tts.dbRepo.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "BeginTransaction")
	}

	defer tx.Rollback()

	token, err := tts.dbRepo.GetNotExpiredToken(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "GetNotExpiredToken")
	}

	if token == nil {

		err := tts.updateToken(ctx)
		if err != nil {
			return errors.Wrap(err, "updateToken")
		}

		err = tts.dbRepo.AddToken(ctx, tx, tts.token)
		if err != nil {
			return errors.Wrap(err, "AddToken")
		}

		if err = tx.Commit(); err != nil {
			return errors.Wrap(err, "Commit")
		}

		return nil
	}

	_, err = tts.twitchOauthClient.TwitchOAuthValidateToken(ctx, *token)
	if err != nil {
		if err.Error() == tokenInvalid {

			err = tts.updateToken(ctx)
			if err != nil {
				return errors.Wrap(err, "updateToken")
			}

			err = tts.dbRepo.AddToken(ctx, tx, tts.token)
			if err != nil {
				return errors.Wrap(err, "AddToken")
			}

			err = tts.dbRepo.SetExpiredToken(ctx, tx, *token)
			if err != nil {
				return errors.Wrap(err, "SetExpiredToken")
			}

			if err = tx.Commit(); err != nil {
				return errors.Wrap(err, "Commit")
			}

			return nil
		}

		return errors.Wrap(err, "TwitchOAuthValidateToken")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "Commit")
	}

	tts.token = *token

	return nil
}

func (tts *TwitchTokenService) updateToken(ctx context.Context) error {
	tokenInfo, err := tts.twitchOauthClient.TwitchOAuthGetToken(ctx, "")
	if err != nil {
		return errors.Wrap(err, "get twitch client token")
	}

	if tokenInfo == nil {
		return errors.Wrap(errors.New("empty client resp"), "TwitchOAuthGetToken")
	}

	tts.token = tokenInfo.AccessToken
	return nil
}

// TODO: сервис прекращает работу без оиждания завершения
func (tts *TwitchTokenService) SyncBg(ctx context.Context, updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("stoping bg %s process", twitchTokeCheckBGSync)
			return
		case <-ticker.C:
			logrus.Infof("started bg %s process", twitchTokeCheckBGSync)
			err := tts.Sync(ctx)
			if err != nil {
				logrus.Infof("could not check twitch token: %v", err)
				continue
			}
			logrus.Info("twitch token check was complited")
		}
	}
}
