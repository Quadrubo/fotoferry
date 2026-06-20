package ferry

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0644))

	got, err := hashFile(path)
	require.NoError(t, err)
	want := sha256.Sum256([]byte("hello"))
	assert.Equal(t, hex.EncodeToString(want[:]), got)
}

func TestHashFile_Missing(t *testing.T) {
	_, err := hashFile(filepath.Join(t.TempDir(), "nope"))
	assert.Error(t, err)
}

func TestCopyFile_CreatesParents(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0644))
	dest := filepath.Join(dir, "a/b/c/out")

	require.NoError(t, copyFile(src, dest, 0644, 0755, -1, -1))
	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "data", string(got))
}

func TestCopyFile_AppliesModes(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0600))
	dest := filepath.Join(dir, "sub", "out")

	require.NoError(t, copyFile(src, dest, 0666, 0777, -1, -1))

	fi, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0666), fi.Mode().Perm(), "file mode applied regardless of source mode")

	di, err := os.Stat(filepath.Join(dir, "sub"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0777), di.Mode().Perm(), "created dir mode applied regardless of umask")
}
