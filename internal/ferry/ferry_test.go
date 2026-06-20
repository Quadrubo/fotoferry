package ferry

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/Quadrubo/fotoferry/internal/config"
	"github.com/Quadrubo/fotoferry/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setup(t *testing.T) (*sql.DB, string, string) {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")
	require.NoError(t, os.MkdirAll(src, 0755))
	db, err := store.Open(filepath.Join(dir, "state.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, src, dest
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

func cfgFor(src, dest string) *config.Config {
	return &config.Config{
		Mappings: []config.Mapping{{ID: "m", Source: src, Dest: dest}},
		FileMode: 0644,
		DirMode:  0755,
		OwnerUID: -1,
		OwnerGID: -1,
	}
}

func run(db *sql.DB, cfg *config.Config) Result {
	return Run(context.Background(), db, cfg)
}

func TestRun_CopiesNewFiles(t *testing.T) {
	db, src, dest := setup(t)
	writeFile(t, filepath.Join(src, "2025/10/a.jpg"), "a")
	writeFile(t, filepath.Join(src, "2025/10/b.jpg"), "b")

	r := run(db, cfgFor(src, dest))
	assert.Equal(t, 2, r.Copied)
	assert.Equal(t, 0, r.Errors)

	got, err := os.ReadFile(filepath.Join(dest, "2025/10/a.jpg"))
	require.NoError(t, err)
	assert.Equal(t, "a", string(got))
}

func TestRun_SkipsUnchangedOnRerun(t *testing.T) {
	db, src, dest := setup(t)
	writeFile(t, filepath.Join(src, "a.jpg"), "a")
	cfg := cfgFor(src, dest)

	run(db, cfg)
	r := run(db, cfg)
	assert.Equal(t, 0, r.Copied)
	assert.Equal(t, 1, r.Skipped)
}

func TestRun_DoesNotResurrectDeleted(t *testing.T) {
	db, src, dest := setup(t)
	writeFile(t, filepath.Join(src, "a.jpg"), "a")
	cfg := cfgFor(src, dest)

	run(db, cfg)
	require.NoError(t, os.Remove(filepath.Join(dest, "a.jpg")))

	r := run(db, cfg)
	assert.Equal(t, 0, r.Copied)
	assert.Equal(t, 1, r.Skipped)
	_, err := os.Stat(filepath.Join(dest, "a.jpg"))
	assert.True(t, os.IsNotExist(err), "a file removed from dest must not return")
}

func TestRun_TemplateMoveIsDuplicate(t *testing.T) {
	db, src, dest := setup(t)
	writeFile(t, filepath.Join(src, "2025/10/a.jpg"), "samebytes")
	cfg := cfgFor(src, dest)
	run(db, cfg)

	require.NoError(t, os.MkdirAll(filepath.Join(src, "2025/11"), 0755))
	require.NoError(t, os.Rename(
		filepath.Join(src, "2025/10/a.jpg"),
		filepath.Join(src, "2025/11/a.jpg"),
	))

	r := run(db, cfg)
	assert.Equal(t, 0, r.Copied)
	assert.Equal(t, 1, r.Duplicate)
	_, err := os.Stat(filepath.Join(dest, "2025/11/a.jpg"))
	assert.True(t, os.IsNotExist(err), "moved bytes must not be re-copied under the new path")
}

func TestRun_DryRunWritesNothing(t *testing.T) {
	db, src, dest := setup(t)
	writeFile(t, filepath.Join(src, "a.jpg"), "a")
	cfg := cfgFor(src, dest)
	cfg.DryRun = true

	r := run(db, cfg)
	assert.Equal(t, 1, r.Copied)
	_, err := os.Stat(filepath.Join(dest, "a.jpg"))
	assert.True(t, os.IsNotExist(err))

	cfg.DryRun = false
	r = run(db, cfg)
	assert.Equal(t, 1, r.Copied, "dry-run must not record, so the real run still copies")
}

func TestRun_MissingSourceSkips(t *testing.T) {
	db, _, dest := setup(t)
	cfg := cfgFor(filepath.Join(t.TempDir(), "nonexistent"), dest)

	r := run(db, cfg)
	assert.Equal(t, Result{}, r)
}

func TestRun_IgnoresDotfiles(t *testing.T) {
	db, src, dest := setup(t)
	writeFile(t, filepath.Join(src, ".immich"), "marker")
	writeFile(t, filepath.Join(src, "a.jpg"), "a")

	r := run(db, cfgFor(src, dest))
	assert.Equal(t, 1, r.Copied)
	_, err := os.Stat(filepath.Join(dest, ".immich"))
	assert.True(t, os.IsNotExist(err))
}
