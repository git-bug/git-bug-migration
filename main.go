package main

import (
	"fmt"
	"os"

	mg1b "github.com/MichaelMure/git-bug-migration/migration1/bug"
	mg1i "github.com/MichaelMure/git-bug-migration/migration1/identity"
	mg1r "github.com/MichaelMure/git-bug-migration/migration1/repository"
)

const rootCommandName = "git-bug-migration"

var repo mg1r.ClockedRepo

func main() {
	Migrate01()
}

func Migrate01() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("unable to get the current working directory: %q", err))
	}

	repo, err = mg1r.NewGitRepo(cwd, []mg1r.ClockLoader{mg1b.ClockLoader})
	if err == mg1r.ErrNotARepo {
		panic(fmt.Errorf("%s must be run from within a git repo", rootCommandName))
	}

	identities := []*mg1i.Identity{}
	for streamedIdentity := range mg1i.ReadAllLocalIdentities(repo) {
		if streamedIdentity.Err != nil {
			fmt.Print(fmt.Errorf("Got error when reading identity: %q\n", err))
			continue
		}
		identities = append(identities, streamedIdentity.Identity)
	}

	if err != nil {
		panic(err)
	}

	// Iterating through all the bugs in the repo
	for streamedBug := range mg1b.ReadAllLocalBugs(repo) {
		if streamedBug.Err != nil {
			fmt.Print(fmt.Errorf("Got error when reading bug: %q\n", err))
			continue
		}

		// Getting the old bug
		oldBug := streamedBug.Bug

		b := mg1b.NewBug()
		bugChange := false

		// Iterating over each operation in the bug
		it := mg1b.NewOperationIterator(oldBug)
		for it.Next() {
			operation := it.Value()
			oldAuthor := operation.GetAuthor()

			// Checking if the author is of the legacy (bare) type
			switch oldAuthor.(type) {
			case *mg1i.Bare:
				fmt.Print("Detected legacyAuthor!\n")

				bugChange = true

				// Search existing identities for any traces of this old identity
				var newAuthor *mg1i.Identity = nil
				for _, identity := range identities {
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
				b.Append(operation)
				continue

			// If the author's identity is a new identity type, its fine. Just append it to the cache
			case *mg1i.Identity:
				b.Append(operation)
				continue

			// This should not be reached
			default:
				fmt.Printf("Unknown author type: %T\n", operation.GetAuthor())
			}
		}

		// If the bug has been changed, remove the old bug and commit the new one
		if bugChange {
			err = b.Commit(repo)
			if err != nil {
				fmt.Printf("Got error when attempting to commit new bug: %q\n", err)
			}

			err = mg1b.RemoveLocalBug(repo, oldBug.Id())
			if err != nil {
				fmt.Printf("Got error when attempting to remove bug: %q\n", err)
			}
		}
	}
}
