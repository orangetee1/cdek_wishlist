package migration

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func Up(db *sql.DB, dir string) error {
	goose.SetDialect("postgres")
	return goose.Up(db, dir)
}
