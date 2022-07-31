package twitch_token

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type TwitchTokenService struct {
	db *sqlx.DB
}

func NewTwitchTokenService(db *sqlx.DB) (*TwitchTokenService, error) {
	service := &TwitchTokenService{
		db: db,
	}

	return service, nil
}

func (tts *TwitchTokenService) Sync(ctx context.Context) error {

	return nil
}

func (tts *TwitchTokenService) GetNotExpiredToken(ctx context.Context) (token *string, err error) {

	query := `
		select 
			tt."token" 
		from twitch_tokens tt
		where tt.is_expired = false
		order by tt.created_at 
		limit 1;
	`

	err = tts.db.GetContext(ctx, token, query)

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
		where "token" = '$1';
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
