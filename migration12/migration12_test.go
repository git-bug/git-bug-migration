package migration12

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	mg1b "github.com/MichaelMure/git-bug-migration/migration12/before/bridge/core/auth"
	mg1r "github.com/MichaelMure/git-bug-migration/migration12/before/repository"

	mg2b "github.com/MichaelMure/git-bug-migration/migration12/after/bridge/core/auth"
	mg2r "github.com/MichaelMure/git-bug-migration/migration12/after/repository"
)

func createFolder() (string, error) {
	dir, err := ioutil.TempDir("", "")
	return dir, err
}

func removeFolder(path string) error {
	return os.RemoveAll(path)
}

func TestMigrate12(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err, "got error when attempting to access the current working directory")

	dir, err := createFolder()
	require.NoError(t, err, "got error when creating temporary repository dir with version 1")
	err = os.Chdir(dir)
	require.NoError(t, err, "got error when opening temporary repository folder")

	repo1 := mg1r.NewMockRepoForTest()

	login := mg1b.NewLogin("target", "login")
	loginPassword := mg1b.NewLoginPassword("target", "login", "password")
	token := mg1b.NewToken("target", "value")

	err = mg1b.Store(repo1, login)
	require.NoError(t, err)
	err = mg1b.Store(repo1, loginPassword)
	require.NoError(t, err)
	err = mg1b.Store(repo1, token)
	require.NoError(t, err)

	oldCredentials, err := mg1b.List(repo1)
	require.NoError(t, err)
	require.NoError(t, err)
	sort.Slice(oldCredentials, func(i, j int) bool {
		k := bytes.Compare(oldCredentials[i].Salt(), oldCredentials[j].Salt())
		require.NotEqual(t, 0, k)
		return k < 0
	})

	oldJSON, err := json.Marshal(oldCredentials)
	require.NoError(t, err)

	repo2 := mg2r.NewMockRepoForTest()
	require.NoError(t, err)

	err = Migrate12(repo1, repo2)
	require.NoError(t, err, "got error when migrating repository with version 2")

	newCredentials, err := mg2b.List(repo2)
	require.NoError(t, err)
	sort.Slice(newCredentials, func(i, j int) bool {
		k := bytes.Compare(newCredentials[i].Salt(), newCredentials[j].Salt())
		require.NotEqual(t, 0, k)
		return k < 0
	})

	newJSON, err := json.Marshal(newCredentials)
	require.NoError(t, err)
	require.Equal(t, oldJSON, newJSON)

	err = os.Chdir(cwd)
	err = removeFolder(dir)
	require.NoError(t, err, "got error when removing temporary folder")
}
