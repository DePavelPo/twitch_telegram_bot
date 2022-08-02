package twitch_token

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"

	twitch_oauth_client "twitch_telegram_bot/internal/client/twitch-oauth-client"
)

const (
	twitchTokeCheckBGSync = "twitchTokenCheck_BGSync"
	tokenInvalid          = "token invalid"
)

type TwitchTokenService struct {
	db                *sqlx.DB
	token             string
	twitchOauthClient *twitch_oauth_client.TwitchOauthClient
}

// TODO: прокидывать токен в другие модули
func NewTwitchTokenService(db *sqlx.DB, twitchOauthClient *twitch_oauth_client.TwitchOauthClient) (*TwitchTokenService, error) {
	service := &TwitchTokenService{
		db:                db,
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

	tx, err := tts.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "BeginTxx")
	}

	defer tx.Rollback()

	token, err := tts.GetNotExpiredToken(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "GetNotExpiredToken")
	}

	if token == nil {

		err := tts.updateToken(ctx)
		if err != nil {
			return errors.Wrap(err, "updateToken")
		}

		err = tts.AddToken(ctx, tx, tts.token)
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

			err = tts.AddToken(ctx, tx, tts.token)
			if err != nil {
				return errors.Wrap(err, "AddToken")
			}

			err = tts.SetExpiredToken(ctx, tx, *token)
			if err != nil {
				return errors.Wrap(err, "SetExpiredToken")
			}

			if err = tx.Commit(); err != nil {
				return errors.Wrap(err, "Commit")
			}

			return nil
		}

		return errors.Wrap(err, "TwitchOAuthGetToken")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "Commit")
	}

	tts.token = *token

	return nil
}

func (tts *TwitchTokenService) updateToken(ctx context.Context) error {
	tokenInfo, err := tts.twitchOauthClient.TwitchOAuthGetToken(ctx)
	if err != nil {
		return errors.Wrap(err, "TwitchOAuthGetToken")
	}

	if tokenInfo == nil {
		return errors.Wrap(errors.New("empty client resp"), "TwitchOAuthGetToken")
	}

	tts.token = tokenInfo.Token
	return nil
}

func (tts *TwitchTokenService) GetNotExpiredToken(ctx context.Context, tx *sqlx.Tx) (token *string, err error) {

	query := `
		select 
			tt."token" 
		from twitch_tokens tt
		where tt.is_expired = false
		order by tt.created_at 
		desc
		limit 1;
	`

	err = tx.GetContext(ctx, &token, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}

func (tts *TwitchTokenService) AddToken(ctx context.Context, tx *sqlx.Tx, token string) (err error) {

	query := `
		insert into twitch_tokens ("token") values ($1);
	`

	res, err := tx.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}

func (tts *TwitchTokenService) SetExpiredToken(ctx context.Context, tx *sqlx.Tx, token string) (err error) {

	query := `
		update twitch_tokens 
		set is_expired = true
		where "token" = $1;
	`

	res, err := tx.ExecContext(ctx, query, token)
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
