package migration3

import (
	"errors"
	"fmt"

	afterbug "github.com/MichaelMure/git-bug-migration/migration3/after/bug"
	afterentity "github.com/MichaelMure/git-bug-migration/migration3/after/entity"
	afteridentity "github.com/MichaelMure/git-bug-migration/migration3/after/identity"
	afterrepo "github.com/MichaelMure/git-bug-migration/migration3/after/repository"

	beforebug "github.com/MichaelMure/git-bug-migration/migration3/before/bug"
	beforeentity "github.com/MichaelMure/git-bug-migration/migration3/before/entity"
	beforeidentity "github.com/MichaelMure/git-bug-migration/migration3/before/identity"
	beforerepo "github.com/MichaelMure/git-bug-migration/migration3/before/repository"
)

type Migration3 struct{}

func (m *Migration3) Description() string {
	return "Make bug and identities independent from the storage by making the ID generation self-contained"
}

func (m *Migration3) Run(repoPath string) error {
	oldRepo, err := beforerepo.NewGoGitRepo(repoPath, nil)
	if err != nil {
		return err
	}

	newRepo, err := afterrepo.NewGoGitRepo(repoPath, nil)
	if err != nil {
		return err
	}

	return m.migrate(oldRepo, newRepo)
}

func (m *Migration3) migrate(oldRepo beforerepo.ClockedRepo, newRepo afterrepo.ClockedRepo) error {
	identities := beforeidentity.ReadAllLocal(oldRepo)
	bugs := beforebug.ReadAllLocal(oldRepo)

	migratedIdentities := map[beforeentity.Id]*afteridentity.Identity{}

	for streamedIdentity := range identities {
		if streamedIdentity.Err != nil {
			if errors.Is(streamedIdentity.Err, beforeidentity.ErrInvalidFormatVersion) {
				fmt.Print("skipping bug, already updated\n")
				continue
			} else {
				return streamedIdentity.Err
			}
		}
		oldIdentity := streamedIdentity.Identity
		fmt.Printf("identity %s: ", oldIdentity.Id().Human())
		newIdentity, err := afteridentity.NewIdentityFull(
			newRepo,
			oldIdentity.Name(),
			oldIdentity.Email(),
			oldIdentity.Login(),
			oldIdentity.AvatarUrl(),
			nil,
		)
		if err != nil {
			return err
		}

		migratedIdentities[oldIdentity.Id()] = newIdentity
		if err := newIdentity.Commit(newRepo); err != nil {
			return err
		}
		fmt.Printf("migrated to %s\n", newIdentity.Id().Human())
	}

	for streamedBug := range bugs {
		if streamedBug.Err != nil {
			if errors.Is(streamedBug.Err, beforebug.ErrInvalidFormatVersion) {
				fmt.Print("skipping bug, already updated\n")
				continue
			} else {
				return streamedBug.Err
			}
		}
		oldBug := streamedBug.Bug
		fmt.Printf("bug %s: ", oldBug.Id().Human())
		newBug, err := migrateBug(oldBug, migratedIdentities)
		if err != nil {
			return err
		} else if newBug == nil {
			fmt.Print("skipping bug, already updated\n")
			return nil
		}
		if err := newBug.Commit(newRepo); err != nil {
			return err
		}
		fmt.Printf("migrated to %s\n", newBug.Id().Human())
		if err := beforebug.RemoveBug(oldRepo, oldBug.Id()); err != nil {
			return err
		}
	}

	for oldIdentity := range migratedIdentities {
		if err := beforeidentity.RemoveIdentity(oldRepo, oldIdentity); err != nil {
			return err
		}
	}

	return nil
}

func migrateBug(oldBug *beforebug.Bug, migratedIdentities map[beforeentity.Id]*afteridentity.Identity) (*afterbug.Bug, error) {
	if oldBug.Packs[0].FormatVersion != 2 {
		return nil, nil
	}

	// Making a new bug
	newBug := afterbug.NewBug()

	migratedOperations := map[beforeentity.Id]afterentity.Id{}

	// Iterating over each operation in the bug
	it := beforebug.NewOperationIterator(oldBug)
	for it.Next() {
		oldOperation := it.Value()

		var newOperation afterbug.Operation
		switch operation := oldOperation.(type) {
		case *beforebug.AddCommentOperation:
			newOperation = afterbug.NewAddCommentOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				operation.Message,
				migrateHashes(operation.Files),
			)
		case *beforebug.CreateOperation:
			newOperation = afterbug.NewCreateOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				operation.Title,
				operation.Message,
				migrateHashes(operation.Files),
			)
		case *beforebug.EditCommentOperation:
			newOperation = afterbug.NewEditCommentOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				migratedOperations[operation.Target],
				operation.Message,
				migrateHashes(operation.Files),
			)
		case *beforebug.LabelChangeOperation:
			newOperation = afterbug.NewLabelChangeOperation(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				migrateLabels(operation.Added),
				migrateLabels(operation.Removed),
			)
		case *beforebug.NoOpOperation:
			newOperation = afterbug.NewNoOpOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
			)
		case *beforebug.SetMetadataOperation:
			newOperation = afterbug.NewSetMetadataOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				migratedOperations[operation.Target],
				operation.Metadata,
			)
		case *beforebug.SetStatusOperation:
			newOperation = afterbug.NewSetStatusOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				afterbug.Status(operation.Status),
			)
		case *beforebug.SetTitleOperation:
			newOperation = afterbug.NewSetTitleOp(
				migratedIdentities[operation.Author.Id()],
				operation.Time().Unix(),
				operation.Title,
				operation.Was,
			)
		default:
			return nil, fmt.Errorf("Unknown oldOperation type: %T\n", operation)
		}

		newBug.Append(newOperation)
		migratedOperations[oldOperation.Id()] = newOperation.Id()
	}

	return newBug, nil
}

func migrateHashes(oldHashes []beforerepo.Hash) (newHashes []afterrepo.Hash) {
	for _, hash := range oldHashes {
		newHashes = append(newHashes, afterrepo.Hash(hash))
	}
	return
}

func migrateLabels(oldLabels []beforebug.Label) (newLabels []afterbug.Label) {
	for _, label := range oldLabels {
		newLabels = append(newLabels, afterbug.Label(label))
	}
	return
}
