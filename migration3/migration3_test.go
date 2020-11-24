package migration3

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	afterbug "github.com/MichaelMure/git-bug-migration/migration3/after/bug"
	afterrepo "github.com/MichaelMure/git-bug-migration/migration3/after/repository"
	beforebug "github.com/MichaelMure/git-bug-migration/migration3/before/bug"
	beforeidentity "github.com/MichaelMure/git-bug-migration/migration3/before/identity"
	beforerepo "github.com/MichaelMure/git-bug-migration/migration3/before/repository"
)

func createFolder() (string, error) {
	dir, err := ioutil.TempDir("", "")
	return dir, err
}

func removeFolder(path string) error {
	return os.RemoveAll(path)
}

func TestMigrate23(t *testing.T) {
	cwd, err := os.Getwd()
	require.Nil(t, err, "got error when attempting to access the current working directory")

	var unix = time.Now().Unix()

	dir, err := createFolder()
	require.Nil(t, err, "got error when creating temporary repository dir with version 0")
	err = os.Chdir(dir)
	require.Nil(t, err, "got error when opening temporary repository folder")

	oldRepo, err := beforerepo.InitGitRepo(dir)
	require.Nil(t, err, "got error when initializing old repository")
	newRepo, err := afterrepo.InitGitRepo(dir)
	require.Nil(t, err, "got error when initializing new repository")

	oldVinc := beforeidentity.NewIdentityFull(
		"Vincent Tiu",
		"vincetiu8@gmail.com",
		"invincibot",
		"https://avatars2.githubusercontent.com/u/46623413?s=460&u=56824597898bc22464222f5c33e8eae6d72def5b&v=4",
	)
	err = oldVinc.Commit(oldRepo)
	require.NoError(t, err)

	title := "bug0"
	message := "beep bop bug"
	bug0, _, err := beforebug.Create(oldVinc, unix, title, message)
	require.Nil(t, err, "got error when creating bug")

	err = bug0.Commit(oldRepo)
	require.Nil(t, err, "got error when committing bug")

	m := Migration3{}
	err = m.migrate(oldRepo, newRepo)
	require.Nil(t, err, "got error when migrating repository")

	bugs1 := afterbug.ReadAllLocal(newRepo)
	bug1 := (<-bugs1).Bug

	operations := afterbug.NewOperationIterator(bug1)
	require.Equal(t, true, operations.Next(), "unable to get first operation")

	operation := operations.Value()
	createOperation, ok := operation.(*afterbug.CreateOperation)
	require.True(t, ok)
	require.Equal(t, title, createOperation.Title)
	require.Equal(t, unix, createOperation.UnixTime)
	require.Equal(t, message, createOperation.Message)

	author := operation.GetAuthor()
	require.Equal(t, oldVinc.Name(), author.Name())
	require.Equal(t, oldVinc.Login(), author.Login())
	require.Equal(t, oldVinc.Email(), author.Email())
	require.Equal(t, oldVinc.AvatarUrl(), author.AvatarUrl())

	var bug afterbug.StreamedBug
	require.Equal(t, bug, <-bugs1, "got additional bug when getting bugs in repository")

	err = os.Chdir(cwd)
	err = removeFolder(dir)
	if err != nil {
		fmt.Printf("got error when removing temporary folder: %q", err)
	}
}
