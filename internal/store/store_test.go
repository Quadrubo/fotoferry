package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "state.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestOpen_CreatesDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "state.db")
	db, err := Open(path)
	require.NoError(t, err)
	_ = db.Close()

	_, err = os.Stat(path)
	assert.NoError(t, err, "database file should exist")
}

func TestLookupPath_NotFound(t *testing.T) {
	db := openTestDB(t)
	e, err := LookupPath(db, "m", "a/b.jpg")
	require.NoError(t, err)
	assert.False(t, e.Found)
}

func TestRecordAndLookup(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, Record(db, "m", "hash1", "a/b.jpg", 123, 456))

	e, err := LookupPath(db, "m", "a/b.jpg")
	require.NoError(t, err)
	assert.True(t, e.Found)
	assert.Equal(t, int64(123), e.Size)
	assert.Equal(t, int64(456), e.Mtime)
}

func TestHashExists(t *testing.T) {
	db := openTestDB(t)
	exists, err := HashExists(db, "m", "hash1")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, Record(db, "m", "hash1", "a.jpg", 1, 1))
	exists, err = HashExists(db, "m", "hash1")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestHashExists_ScopedPerMapping(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, Record(db, "m1", "hash1", "a.jpg", 1, 1))

	exists, err := HashExists(db, "m2", "hash1")
	require.NoError(t, err)
	assert.False(t, exists, "a hash recorded under m1 must not match m2")
}

func TestRecord_UpsertSamePath(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, Record(db, "m", "hash1", "a.jpg", 1, 100))
	require.NoError(t, Record(db, "m", "hash2", "a.jpg", 2, 200))

	e, err := LookupPath(db, "m", "a.jpg")
	require.NoError(t, err)
	assert.Equal(t, int64(2), e.Size)
	assert.Equal(t, int64(200), e.Mtime)

	var count int
	require.NoError(t, db.QueryRow(
		"SELECT COUNT(*) FROM copied WHERE mapping = ? AND rel_path = ?", "m", "a.jpg",
	).Scan(&count))
	assert.Equal(t, 1, count)
}
