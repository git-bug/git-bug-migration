package migration01

import (
	"fmt"

	mg1b "github.com/MichaelMure/git-bug-migration/migration01/after/bug"
	mg1i "github.com/MichaelMure/git-bug-migration/migration01/after/identity"
	mg1r "github.com/MichaelMure/git-bug-migration/migration01/after/repository"
)

var identities1 []*mg1i.Identity

func readIdentities(repo mg1r.ClockedRepo) {
	for streamedIdentity := range mg1i.ReadAllLocalIdentities(repo) {
		if streamedIdentity.Err != nil {
			fmt.Printf("Got error when reading identity: %q", streamedIdentity.Err)
			continue
		}
		identities1 = append(identities1, streamedIdentity.Identity)
	}
}

func Migrate01(repo mg1r.ClockedRepo) error {
	readIdentities(repo)

	// Iterating through all the bugs in the repo
	for streamedBug := range mg1b.ReadAllLocalBugs(repo) {
		if streamedBug.Err != nil {
			fmt.Printf("Got error when reading bug: %q\n", streamedBug.Err)
			continue
		}

		oldBug := streamedBug.Bug
		newBug, changed, err := migrateBug01(oldBug)
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

			err = mg1b.RemoveLocalBug(repo, oldBug.Id())
			if err != nil {
				fmt.Printf("Got error when attempting to remove bug: %q\n", err)
				continue
			}
		}
	}

	return nil
}

func migrateBug01(oldBug *mg1b.Bug) (*mg1b.Bug, bool, error) {
	// Making a new bug
	newBug := mg1b.NewBug()
	bugChange := false

	// Iterating over each operation in the bug
	it := mg1b.NewOperationIterator(oldBug)
	for it.Next() {
		operation := it.Value()
		oldAuthor := operation.GetAuthor()

		// Checking if the author is of the legacy (bare) type
		switch oldAuthor.(type) {
		case *mg1i.Bare:
			bugChange = true

			// Search existing identities for any traces of this old identity
			var newAuthor *mg1i.Identity = nil
			for _, identity := range identities1 {
				if oldAuthor.Name() == identity.Name() {
					newAuthor = identity
				}
			}

			// If no existing identity is found, create a new one
			if newAuthor == nil {
				newAuthor = mg1i.NewIdentityFull(
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
		case *mg1i.Identity:
			newBug.Append(operation)
			continue

		// This should not be reached
		default:
			return newBug, false, fmt.Errorf("Unknown author type: %T\n", operation.GetAuthor())
		}
	}

	return newBug, bugChange, nil
}
