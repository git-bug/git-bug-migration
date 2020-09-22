package migration1

import (
	"fmt"

	afterbug "github.com/MichaelMure/git-bug-migration/migration1/after/bug"
	afteridentity "github.com/MichaelMure/git-bug-migration/migration1/after/identity"
	afterrepo "github.com/MichaelMure/git-bug-migration/migration1/after/repository"
)

type Migration1 struct {
	allIdentities []*afteridentity.Identity
}

func (m *Migration1) Description() string {
	return "Convert legacy identities into a complete data structure in git"
}

func (m *Migration1) Run(repoPath string) error {
	repo, err := afterrepo.NewGitRepo(repoPath, nil)
	if err != nil {
		return err
	}

	return m.migrate(repo)
}

func (m *Migration1) migrate(repo afterrepo.ClockedRepo) error {
	m.readIdentities(repo)

	// Iterating through all the bugs in the repo
	for streamedBug := range afterbug.ReadAllLocalBugs(repo) {
		if streamedBug.Err != nil {
			fmt.Printf("Got error when reading bug: %q\n", streamedBug.Err)
			continue
		}

		oldBug := streamedBug.Bug
		newBug, changed, err := m.migrateBug(oldBug)
		if err != nil {
			fmt.Printf("Got error when parsing bug: %q", err)
		}

		// If the bug has been changed, remove the old bug and commit the new one
		if changed {
			err = newBug.Commit(repo)
			if err != nil {
				fmt.Printf("Got error when attempting to commit new bug: %q\n", err)
				continue
			}

			err = afterbug.RemoveLocalBug(repo, oldBug.Id())
			if err != nil {
				fmt.Printf("Got error when attempting to remove bug: %q\n", err)
				continue
			}
		}
	}

	return nil
}

func (m *Migration1) readIdentities(repo afterrepo.ClockedRepo) {
	for streamedIdentity := range afteridentity.ReadAllLocalIdentities(repo) {
		if streamedIdentity.Err != nil {
			fmt.Printf("Got error when reading identity: %q", streamedIdentity.Err)
			continue
		}
		m.allIdentities = append(m.allIdentities, streamedIdentity.Identity)
	}
}

func (m *Migration1) migrateBug(oldBug *afterbug.Bug) (*afterbug.Bug, bool, error) {
	// Making a new bug
	newBug := afterbug.NewBug()
	bugChange := false

	// Iterating over each operation in the bug
	it := afterbug.NewOperationIterator(oldBug)
	for it.Next() {
		operation := it.Value()
		oldAuthor := operation.GetAuthor()

		// Checking if the author is of the legacy (bare) type
		switch oldAuthor.(type) {
		case *afteridentity.Bare:
			bugChange = true

			// Search existing identities for any traces of this old identity
			var newAuthor *afteridentity.Identity = nil
			for _, identity := range m.allIdentities {
				if oldAuthor.Name() == identity.Name() {
					newAuthor = identity
				}
			}

			// If no existing identity is found, create a new one
			if newAuthor == nil {
				newAuthor = afteridentity.NewIdentityFull(
					oldAuthor.Name(),
					oldAuthor.Email(),
					oldAuthor.Login(),
					oldAuthor.AvatarUrl(),
				)
			}

			// Set the author of the operation to the new identity
			operation.SetAuthor(newAuthor)
			newBug.Append(operation)
			continue

		// If the author's identity is a new identity type, its fine. Just append it to the cache
		case *afteridentity.Identity:
			newBug.Append(operation)
			continue

		// This should not be reached
		default:
			return newBug, false, fmt.Errorf("Unknown author type: %T\n", operation.GetAuthor())
		}
	}

	return newBug, bugChange, nil
}
