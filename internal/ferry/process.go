package ferry

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/Quadrubo/fotoferry/internal/config"
	"github.com/Quadrubo/fotoferry/internal/store"
)

type outcome int

const (
	outcomeSkipped outcome = iota
	outcomeCopied
	outcomeDuplicate
)

func processFile(db *sql.DB, cfg *config.Config, m config.Mapping, srcPath, rel string, size, mtime int64) (outcome, error) {
	seen, err := store.LookupPath(db, m.ID, rel)
	if err != nil {
		return 0, fmt.Errorf("lookup: %w", err)
	}
	if seen.Found && seen.Size == size && seen.Mtime == mtime {
		return outcomeSkipped, nil
	}

	sha, err := hashFile(srcPath)
	if err != nil {
		return 0, fmt.Errorf("hash: %w", err)
	}

	// Same bytes already delivered (e.g. a storage-template move): record the path, don't copy.
	exists, err := store.HashExists(db, m.ID, sha)
	if err != nil {
		return 0, fmt.Errorf("hash lookup: %w", err)
	}
	if exists {
		if !cfg.DryRun {
			if err := store.Record(db, m.ID, sha, rel, size, mtime); err != nil {
				return 0, fmt.Errorf("record: %w", err)
			}
		}
		return outcomeDuplicate, nil
	}

	if cfg.DryRun {
		return outcomeCopied, nil
	}
	if err := copyFile(srcPath, filepath.Join(m.Dest, rel), cfg.FileMode, cfg.DirMode, cfg.OwnerUID, cfg.OwnerGID); err != nil {
		return 0, fmt.Errorf("copy: %w", err)
	}
	if err := store.Record(db, m.ID, sha, rel, size, mtime); err != nil {
		return 0, fmt.Errorf("record: %w", err)
	}
	return outcomeCopied, nil
}
