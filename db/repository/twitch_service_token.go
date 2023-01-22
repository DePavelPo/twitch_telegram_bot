package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

func (dbr *DBRepository) GetNotExpiredToken(ctx context.Context, tx *sqlx.Tx) (token *string, err error) {

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

func (dbr *DBRepository) AddToken(ctx context.Context, tx *sqlx.Tx, token string) (err error) {

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

func (dbr *DBRepository) SetExpiredToken(ctx context.Context, tx *sqlx.Tx, token string) (err error) {

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
