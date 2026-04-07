package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes a connection to a SQLite database file.
func InitDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	initPath := "./internal/db/init/"
	info, err := os.Stat(initPath)
	if err == nil && info.IsDir() {
		files, err := os.ReadDir(initPath)
		if err != nil {
			return db, fmt.Errorf("failed to open init file: %w", err)
		}
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(initPath, file.Name()))
			if err != nil {
				return db, fmt.Errorf("failed to read init file: %w", err)
			}
			_, err = db.Exec(string(content))
			if err != nil {
				return db, fmt.Errorf("failed to apply sqlite schema: %w", err)
			}
		}
	}

	return db, nil
}
