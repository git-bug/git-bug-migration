package migration2

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	afterauth "github.com/MichaelMure/git-bug-migration/migration2/after/bridge/core/auth"
	afterrepo "github.com/MichaelMure/git-bug-migration/migration2/after/repository"
	beforeauth "github.com/MichaelMure/git-bug-migration/migration2/before/bridge/core/auth"
	beforerepo "github.com/MichaelMure/git-bug-migration/migration2/before/repository"
)

func TestMigrate12(t *testing.T) {
	repo1 := beforerepo.NewMockRepoForTest()

	login := beforeauth.NewLogin("target", "login")
	loginPassword := beforeauth.NewLoginPassword("target", "login", "password")
	token := beforeauth.NewToken("target", "value")

	err := beforeauth.Store(repo1, login)
	require.NoError(t, err)
	err = beforeauth.Store(repo1, loginPassword)
	require.NoError(t, err)
	err = beforeauth.Store(repo1, token)
	require.NoError(t, err)

	repo2 := afterrepo.NewMockRepoForTest()
	require.NoError(t, err)

	m := Migration2{}
	err = m.migrate(repo1, repo2)
	require.NoError(t, err, "got error when migrating repository with version 2")

	// to avoid the types problem, we compare the json output of those credentials
	// which also mean we need to ensure a deterministic ordering:

	oldCredentials, err := beforeauth.List(repo1)
	require.NoError(t, err)
	sort.Slice(oldCredentials, func(i, j int) bool {
		k := bytes.Compare(oldCredentials[i].Salt(), oldCredentials[j].Salt())
		require.NotEqual(t, 0, k)
		return k < 0
	})

	oldJSON, err := json.Marshal(oldCredentials)
	require.NoError(t, err)

	newCredentials, err := afterauth.List(repo2)
	require.NoError(t, err)
	sort.Slice(newCredentials, func(i, j int) bool {
		k := bytes.Compare(newCredentials[i].Salt(), newCredentials[j].Salt())
		require.NotEqual(t, 0, k)
		return k < 0
	})

	newJSON, err := json.Marshal(newCredentials)
	require.NoError(t, err)
	require.Equal(t, oldJSON, newJSON)
}
