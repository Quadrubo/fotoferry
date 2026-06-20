package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setMapping(t *testing.T) {
	t.Helper()
	t.Setenv("MAPPING__0__ID", "alice")
	t.Setenv("MAPPING__0__SOURCE", "/library/alice")
	t.Setenv("MAPPING__0__DEST", "/dest/alice")
}

func TestLoad_Defaults(t *testing.T) {
	setMapping(t)
	cfg, err := Load("")
	require.NoError(t, err)

	require.Len(t, cfg.Mappings, 1)
	assert.Equal(t, "/data/state.db", cfg.StateDB)
	assert.Equal(t, "text", cfg.LogFormat)
	assert.False(t, cfg.DryRun)
	assert.Nil(t, cfg.RequirePaths)
	assert.Equal(t, os.FileMode(0644), cfg.FileMode)
	assert.Equal(t, os.FileMode(0755), cfg.DirMode)
	assert.Equal(t, -1, cfg.OwnerUID)
	assert.Equal(t, -1, cfg.OwnerGID)
}

func TestLoad_OwnerAndModes(t *testing.T) {
	setMapping(t)
	t.Setenv("FILE_MODE", "0666")
	t.Setenv("DIR_MODE", "0777")
	t.Setenv("OWNER_UID", "1000")
	t.Setenv("OWNER_GID", "100")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0666), cfg.FileMode)
	assert.Equal(t, os.FileMode(0777), cfg.DirMode)
	assert.Equal(t, 1000, cfg.OwnerUID)
	assert.Equal(t, 100, cfg.OwnerGID)
}

func TestLoad_Overrides(t *testing.T) {
	setMapping(t)
	t.Setenv("STATE_DB", "/tmp/s.db")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("DRY_RUN", "true")
	t.Setenv("REQUIRE_PATHS", "/dest, /other")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/s.db", cfg.StateDB)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, []string{"/dest", "/other"}, cfg.RequirePaths)
}

func TestLoad_MissingMappings(t *testing.T) {
	cfg, err := Load("")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_MappingMissingDest(t *testing.T) {
	t.Setenv("MAPPING__0__ID", "k")
	t.Setenv("MAPPING__0__SOURCE", "/s")

	_, err := Load("")
	assert.Error(t, err)
}
