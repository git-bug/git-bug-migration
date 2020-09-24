package migration2

import (
	"errors"

	afterauth "github.com/MichaelMure/git-bug-migration/migration2/after/bridge/core/auth"
	afterentity "github.com/MichaelMure/git-bug-migration/migration2/after/entity"
	afterrepo "github.com/MichaelMure/git-bug-migration/migration2/after/repository"
	beforeauth "github.com/MichaelMure/git-bug-migration/migration2/before/bridge/core/auth"
	beforerepo "github.com/MichaelMure/git-bug-migration/migration2/before/repository"
)

type Migration2 struct{}

func (m *Migration2) Description() string {
	return "Migrate bridge credentials from the global git config to a keyring"
}

func (m *Migration2) Run(repoPath string) error {
	oldRepo, err := beforerepo.NewGitRepo(repoPath, func(beforerepo.ClockedRepo) error { return nil })
	if err != nil {
		return err
	}

	newRepo, err := afterrepo.NewGitRepo(repoPath, nil)
	if err != nil {
		return err
	}

	return m.migrate(oldRepo, newRepo)
}

func (m *Migration2) migrate(oldRepo beforerepo.ClockedRepo, newRepo afterrepo.ClockedRepo) error {
	creds, err := beforeauth.List(oldRepo)
	if err != nil {
		return err
	}

	for _, cred := range creds {
		var newCred afterauth.Credential
		if afterauth.IdExist(newRepo, afterentity.Id(cred.ID().String())) {
			continue
		}

		switch cred := cred.(type) {
		case *beforeauth.Login:
			newCred = &afterauth.Login{
				Login: cred.Login,
				CredentialBase: afterauth.NewCredentialBase(
					cred.Target(),
					cred.CreateTime(),
					cred.Salt(),
					cred.Metadata(),
				),
			}
		case *beforeauth.LoginPassword:
			newCred = &afterauth.LoginPassword{
				Login:    cred.Login,
				Password: cred.Password,
				CredentialBase: afterauth.NewCredentialBase(
					cred.Target(),
					cred.CreateTime(),
					cred.Salt(),
					cred.Metadata(),
				),
			}
		case *beforeauth.Token:
			newCred = &afterauth.Token{
				Value: cred.Value,
				CredentialBase: afterauth.NewCredentialBase(
					cred.Target(),
					cred.CreateTime(),
					cred.Salt(),
					cred.Metadata(),
				),
			}
		default:
			return errors.New("unknown credential encountered")
		}

		err = afterauth.Store(newRepo, newCred)
		if err != nil {
			return err
		}
	}

	return nil
}
