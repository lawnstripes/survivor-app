package db

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

//go:embed init/*.sql
var initFS embed.FS

// InitDB initializes a connection to a SQLite database file.
func InitDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	files, err := initFS.ReadDir("init")
	if err != nil {
		return db, fmt.Errorf("failed to read embedded init directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := initFS.ReadFile(filepath.Join("init", file.Name()))
		if err != nil {
			return db, fmt.Errorf("failed to read embedded init file %s: %w", file.Name(), err)
		}
		_, err = db.Exec(string(content))
		if err != nil {
			return db, fmt.Errorf("failed to apply sqlite schema from %s: %w", file.Name(), err)
		}
	}

	return db, nil
}
