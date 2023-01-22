package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func (dbr *DBRepository) BeginTransaction(ctx context.Context) (tx *sqlx.Tx, err error) {

	return dbr.db.BeginTxx(ctx, &sql.TxOptions{})

}
