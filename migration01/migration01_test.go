package migration01

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	mg1b "github.com/MichaelMure/git-bug-migration/migration01/after/bug"
	mg1i "github.com/MichaelMure/git-bug-migration/migration01/after/identity"
	mg0b "github.com/MichaelMure/git-bug-migration/migration01/before/bug"
	mg0r "github.com/MichaelMure/git-bug-migration/migration01/before/repository"

	mg1r "github.com/MichaelMure/git-bug-migration/migration01/after/repository"
)

func createFolder() (string, error) {
	dir, err := ioutil.TempDir("", "")
	return dir, err
}

func removeFolder(path string) error {
	return os.RemoveAll(path)
}

func TestMigrate01(t *testing.T) {
	cwd, err := os.Getwd()
	require.Nil(t, err, "got error when attempting to access the current working directory")

	oldVinc := mg0b.Person{
		Name:      "Vincent Tiu",
		Email:     "vincetiu8@gmail.com",
		Login:     "invincibot",
		AvatarUrl: "https://avatars2.githubusercontent.com/u/46623413?s=460&u=56824597898bc22464222f5c33e8eae6d72def5b&v=4",
	}

	newVinc := mg1i.NewIdentityFull(
		oldVinc.Name,
		oldVinc.Email,
		oldVinc.Login,
		oldVinc.AvatarUrl,
	)

	var unix = time.Now().Unix()

	dir, err := createFolder()
	require.Nil(t, err, "got error when creating temporary repository dir with version 0")
	err = os.Chdir(dir)
	require.Nil(t, err, "got error when opening temporary repository folder")

	repo0, err := mg0r.InitGitRepo(dir)

	bug0, _, err := mg0b.Create(oldVinc, unix, "bug1", "beep bop bug")
	require.Nil(t, err, "got error when creating bug")

	err = bug0.Commit(repo0)
	require.Nil(t, err, "got error when committing bug")

	repo1, err := mg1r.NewGitRepo(dir, []mg1r.ClockLoader{mg1b.ClockLoader})
	require.Nil(t, err, "got error when loading repository with version 1")

	err = Migrate01(repo1)
	require.Nil(t, err, "got error when migrating repository with version 1")

	bugs1 := mg1b.ReadAllLocalBugs(repo1)
	bug1 := (<-bugs1).Bug
	operations := mg1b.NewOperationIterator(bug1)
	require.Equal(t, true, operations.Next(), "unable to get first operation")

	operation := operations.Value()

	author := operation.GetAuthor()
	require.IsType(t, newVinc, author, "author type mismatch")
	require.Equal(t, newVinc.Name(), author.Name(), "author name mismatch")
	require.Equal(t, newVinc.Email(), author.Email(), "author email mismatch")
	require.Equal(t, newVinc.Login(), author.Login(), "author login mismatch")
	require.Equal(t, newVinc.AvatarUrl(), author.AvatarUrl(), "author avatarUrl mismatch")

	var bug mg1b.StreamedBug
	require.Equal(t, bug, <-bugs1, "got additional bug when getting bugs in repository")

	err = os.Chdir(cwd)
	err = removeFolder(dir)
	if err != nil {
		fmt.Printf("got error when removing temporary folder: %q", err)
	}
}
