package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGroup_Mappings(t *testing.T) {
	v := viper.New()
	v.Set("MAPPING__0__ID", "alice")
	v.Set("MAPPING__0__SOURCE", "/library/alice")
	v.Set("MAPPING__0__DEST", "/dest/alice")
	v.Set("MAPPING__1__ID", "bob")
	v.Set("MAPPING__1__SOURCE", "/library/bob")
	v.Set("MAPPING__1__DEST", "/dest/bob")

	mappings, err := parseGroup[Mapping](v, "MAPPING", "ID")
	require.NoError(t, err)

	require.Len(t, mappings, 2)
	assert.Equal(t, "alice", mappings[0].ID)
	assert.Equal(t, "/library/bob", mappings[1].Source)
	assert.Equal(t, "/dest/bob", mappings[1].Dest)
}

func TestParseGroup_StopsAtGap(t *testing.T) {
	v := viper.New()
	v.Set("MAPPING__0__ID", "a")
	v.Set("MAPPING__0__SOURCE", "/s")
	v.Set("MAPPING__0__DEST", "/d")
	v.Set("MAPPING__2__ID", "c")

	mappings, err := parseGroup[Mapping](v, "MAPPING", "ID")
	require.NoError(t, err)
	assert.Len(t, mappings, 1, "should stop at the gap in indices")
}

func TestSplitList(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitList("a, b ,c"))
	assert.Nil(t, splitList(""))
	assert.Nil(t, splitList("  ,  "))
}
