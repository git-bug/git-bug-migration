package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T, config Config) {
	err := config.StoreString("section.key", "value")
	require.NoError(t, err)

	val, err := config.ReadString("section.key")
	require.NoError(t, err)
	require.Equal(t, "value", val)

	err = config.StoreString("section.true", "true")
	require.NoError(t, err)

	val2, err := config.ReadBool("section.true")
	require.NoError(t, err)
	require.Equal(t, true, val2)

	configs, err := config.ReadAll("section")
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"section.key":  "value",
		"section.true": "true",
	}, configs)

	err = config.RemoveAll("section.true")
	require.NoError(t, err)

	configs, err = config.ReadAll("section")
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"section.key": "value",
	}, configs)

	_, err = config.ReadBool("section.true")
	require.Equal(t, ErrNoConfigEntry, err)

	err = config.RemoveAll("section.nonexistingkey")
	require.Error(t, err)

	err = config.RemoveAll("section.key")
	require.NoError(t, err)

	_, err = config.ReadString("section.key")
	require.Equal(t, ErrNoConfigEntry, err)

	err = config.RemoveAll("nonexistingsection")
	require.Error(t, err)

	err = config.RemoveAll("section")
	require.Error(t, err)

	_, err = config.ReadString("section.key")
	require.Error(t, err)

	err = config.RemoveAll("section.key")
	require.Error(t, err)
}
