package storage

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	Db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.mysql.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			telegram TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_telegram ON users(telegram)")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := Storage{
		Db: db,
	}

	return &storage, nil
}
