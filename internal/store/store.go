package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"github.com/Quadrubo/fotoferry/internal/migrations"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=wal"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrations.Run(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

type Entry struct {
	Size  int64
	Mtime int64
	Found bool
}

func LookupPath(db *sql.DB, mapping, relPath string) (Entry, error) {
	var e Entry
	err := db.QueryRow(
		"SELECT size, mtime FROM copied WHERE mapping = ? AND rel_path = ?",
		mapping, relPath,
	).Scan(&e.Size, &e.Mtime)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, nil
	}
	if err != nil {
		return Entry{}, err
	}
	e.Found = true
	return e, nil
}

func HashExists(db *sql.DB, mapping, sha string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM copied WHERE mapping = ? AND sha256 = ?)",
		mapping, sha,
	).Scan(&exists)
	return exists, err
}

func Record(db *sql.DB, mapping, sha, relPath string, size, mtime int64) error {
	_, err := db.Exec(`
		INSERT INTO copied (mapping, sha256, rel_path, size, mtime)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(mapping, rel_path) DO UPDATE SET
			sha256    = excluded.sha256,
			size      = excluded.size,
			mtime     = excluded.mtime,
			copied_at = CURRENT_TIMESTAMP
	`, mapping, sha, relPath, size, mtime)
	return err
}
