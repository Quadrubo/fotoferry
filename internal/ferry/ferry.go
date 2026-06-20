package ferry

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Quadrubo/fotoferry/internal/config"
)

type Result struct {
	Copied    int
	Skipped   int
	Duplicate int
	Errors    int
}

func Run(ctx context.Context, db *sql.DB, cfg *config.Config) Result {
	var total Result
	for _, m := range cfg.Mappings {
		r := runMapping(ctx, db, cfg, m)
		total.Copied += r.Copied
		total.Skipped += r.Skipped
		total.Duplicate += r.Duplicate
		total.Errors += r.Errors
	}
	return total
}

func runMapping(ctx context.Context, db *sql.DB, cfg *config.Config, m config.Mapping) Result {
	var r Result

	if fi, err := os.Stat(m.Source); err != nil || !fi.IsDir() {
		slog.Warn("source missing, skipping mapping", "mapping", m.ID, "source", m.Source)
		return r
	}

	walkErr := filepath.WalkDir(m.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("walk error", "mapping", m.ID, "path", path, "error", err)
			r.Errors++
			return nil
		}
		if d.IsDir() || !d.Type().IsRegular() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		rel, err := filepath.Rel(m.Source, path)
		if err != nil {
			slog.Error("rel path", "mapping", m.ID, "path", path, "error", err)
			r.Errors++
			return nil
		}
		info, err := d.Info()
		if err != nil {
			slog.Error("stat", "mapping", m.ID, "file", rel, "error", err)
			r.Errors++
			return nil
		}

		out, err := processFile(db, cfg, m, path, rel, info.Size(), info.ModTime().Unix())
		if err != nil {
			slog.Error("process file", "mapping", m.ID, "file", rel, "error", err)
			r.Errors++
			return nil
		}
		switch out {
		case outcomeSkipped:
			r.Skipped++
		case outcomeDuplicate:
			r.Duplicate++
		case outcomeCopied:
			r.Copied++
			if cfg.DryRun {
				slog.Info("would copy", "mapping", m.ID, "file", rel)
			} else {
				slog.Info("copied", "mapping", m.ID, "file", rel)
			}
		}
		return nil
	})

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		slog.Error("walk failed", "mapping", m.ID, "error", walkErr)
		r.Errors++
	}
	slog.Info("mapping done", "mapping", m.ID,
		"copied", r.Copied, "skipped", r.Skipped, "duplicate", r.Duplicate, "errors", r.Errors)
	return r
}
