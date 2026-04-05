package payment

import "github.com/jmoiron/sqlx"

type Repository struct {
	db *sqlx.DB
}

func New(d *sqlx.DB) *Repository {
	return &Repository{
		db: d,
	}
}
