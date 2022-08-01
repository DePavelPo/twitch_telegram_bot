package twitch_token

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"

	twitch_client "twitch_telegram_bot/internal/client/twitch-client"
)

const (
	twitchTokeCheckBGSync = "twitchTokenCheck_BGSync"
	tokenInvalid          = "token invalid"
)

type TwitchTokenService struct {
	db           *sqlx.DB
	token        string
	twitchClient *twitch_client.TwitchClient
}

// TODO: прокидывать токен в другие модули
func NewTwitchTokenService(db *sqlx.DB, twitchClient *twitch_client.TwitchClient) (*TwitchTokenService, error) {
	service := &TwitchTokenService{
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

// TODO: подумать над оптимизацией
func (tts *TwitchTokenService) Sync(ctx context.Context) error {

	token, err := tts.GetNotExpiredToken(ctx)
	if err != nil {
		return errors.Wrap(err, "GetNotExpiredToken")
	}

	if token == nil {

		// TODO: завернуть в транзакцию
		err := tts.updateToken(ctx)
		if err != nil {
			return errors.Wrap(err, "updateToken")
		}

		err = tts.AddToken(ctx, tts.token)
		if err != nil {
			return errors.Wrap(err, "AddToken")
		}

		return nil
	}

	_, err = tts.twitchClient.TwitchOAuthValidateToken(ctx, *token)
	if err != nil {
		if err.Error() == tokenInvalid {

			// TODO: завернуть в транзакцию
			err = tts.updateToken(ctx)
			if err != nil {
				return errors.Wrap(err, "updateToken")
			}

			err = tts.AddToken(ctx, tts.token)
			if err != nil {
				return errors.Wrap(err, "AddToken")
			}

			err = tts.SetExpiredToken(ctx, *token)
			if err != nil {
				return errors.Wrap(err, "SetExpiredToken")
			}

			return nil
		}

		return errors.Wrap(err, "TwitchOAuthGetToken")
	}

	return nil
}

func (tts *TwitchTokenService) updateToken(ctx context.Context) error {
	tokenInfo, err := tts.twitchClient.TwitchOAuthGetToken(ctx)
	if err != nil {
		return errors.Wrap(err, "TwitchOAuthGetToken")
	}

	if tokenInfo == nil {
		return errors.Wrap(errors.New("empty client resp"), "TwitchOAuthGetToken")
	}

	tts.token = tokenInfo.Token
	return nil
}

func (tts *TwitchTokenService) GetNotExpiredToken(ctx context.Context) (token *string, err error) {

	query := `
		select 
			tt."token" 
		from twitch_tokens tt
		where tt.is_expired = false
		order by tt.created_at 
		desc
		limit 1;
	`

	err = tts.db.GetContext(ctx, &token, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}

func (tts *TwitchTokenService) AddToken(ctx context.Context, token string) (err error) {

	query := `
		insert into twitch_tokens ("token") values ($1);
	`

	res, err := tts.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}

func (tts *TwitchTokenService) SetExpiredToken(ctx context.Context, token string) (err error) {

	query := `
		update twitch_tokens 
		set is_expired = true
		where "token" = $1;
	`

	res, err := tts.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n < 1 {
		return errors.New("token not found")
	}

	return
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
