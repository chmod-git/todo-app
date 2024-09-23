package repository

import "github.com/jmoiron/sqlx"

type ListSQL struct {
	db *sqlx.DB
}
