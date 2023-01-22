package repository

import "github.com/jmoiron/sqlx"

type DBRepository struct {
	db *sqlx.DB
}

func NewDBRepository(db *sqlx.DB) *DBRepository {
	return &DBRepository{
		db: db,
	}
}
