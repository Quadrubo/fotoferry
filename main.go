package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Quadrubo/fotoferry/internal/config"
	"github.com/Quadrubo/fotoferry/internal/ferry"
	"github.com/Quadrubo/fotoferry/internal/store"
	_ "modernc.org/sqlite"
)

func main() {
	envFile := flag.String("env-file", "", "path to .env config file (default: read from environment)")
	flag.Parse()

	cfg, err := config.Load(*envFile)
	if err != nil {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	setupLogger(cfg.LogFormat)

	for _, p := range cfg.RequirePaths {
		if _, err := os.Stat(p); err != nil {
			slog.Warn("required path missing, skipping run", "path", p)
			return
		}
	}

	db, err := store.Open(cfg.StateDB)
	if err != nil {
		slog.Error("failed to open state db", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	unlock, err := acquireLock(cfg.StateDB)
	if err != nil {
		slog.Warn("another run is in progress, exiting")
		return
	}
	defer unlock()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.DryRun {
		slog.Info("dry run, no files will be copied")
	}

	r := ferry.Run(ctx, db, cfg)
	slog.Info("run complete",
		"copied", r.Copied, "skipped", r.Skipped, "duplicate", r.Duplicate, "errors", r.Errors)
	if r.Errors > 0 {
		os.Exit(1)
	}
}

func setupLogger(format string) {
	var h slog.Handler
	if format == "json" {
		h = slog.NewJSONHandler(os.Stderr, nil)
	} else {
		h = slog.NewTextHandler(os.Stderr, nil)
	}
	slog.SetDefault(slog.New(h))
}

func acquireLock(stateDB string) (func(), error) {
	path := filepath.Join(filepath.Dir(stateDB), ".fotoferry.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}
