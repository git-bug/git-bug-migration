// Package repository contains helper methods for working with the Git repo.
package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	repo := CreateTestRepo(false)
	defer CleanupTestRepos(t, repo)

	err := repo.LocalConfig().StoreString("section.key", "value")
	require.NoError(t, err)

	val, err := repo.LocalConfig().ReadString("section.key")
	require.NoError(t, err)
	require.Equal(t, "value", val)

	err = repo.LocalConfig().StoreString("section.true", "true")
	require.NoError(t, err)

	val2, err := repo.LocalConfig().ReadBool("section.true")
	require.NoError(t, err)
	require.Equal(t, true, val2)

	configs, err := repo.LocalConfig().ReadAll("section")
	require.NoError(t, err)
	require.Equal(t, configs, map[string]string{
		"section.key":  "value",
		"section.true": "true",
	})

	err = repo.LocalConfig().RemoveAll("section.true")
	require.NoError(t, err)

	configs, err = repo.LocalConfig().ReadAll("section")
	require.NoError(t, err)
	require.Equal(t, configs, map[string]string{
		"section.key": "value",
	})

	_, err = repo.LocalConfig().ReadBool("section.true")
	require.Equal(t, ErrNoConfigEntry, err)

	err = repo.LocalConfig().RemoveAll("section.nonexistingkey")
	require.Error(t, err)

	err = repo.LocalConfig().RemoveAll("section.key")
	require.NoError(t, err)

	_, err = repo.LocalConfig().ReadString("section.key")
	require.Equal(t, ErrNoConfigEntry, err)

	err = repo.LocalConfig().RemoveAll("nonexistingsection")
	require.Error(t, err)

	err = repo.LocalConfig().RemoveAll("section")
	require.Error(t, err)

	_, err = repo.LocalConfig().ReadString("section.key")
	require.Error(t, err)

	err = repo.LocalConfig().RemoveAll("section.key")
	require.Error(t, err)
}
