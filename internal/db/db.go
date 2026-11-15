package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

// InitDB initializes a connection to a SQLite database file.
func InitDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}
	return db, nil
}
