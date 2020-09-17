package migration12

import (
	"errors"

	mg2b "github.com/MichaelMure/git-bug-migration/migration12/after/bridge/core/auth"
	mg2e "github.com/MichaelMure/git-bug-migration/migration12/after/entity"
	mg2r "github.com/MichaelMure/git-bug-migration/migration12/after/repository"
	mg1b "github.com/MichaelMure/git-bug-migration/migration12/before/bridge/core/auth"
	mg1r "github.com/MichaelMure/git-bug-migration/migration12/before/repository"
)

func Migrate12(oldRepo mg1r.ClockedRepo, newRepo mg2r.ClockedRepo) error {
	creds, err := mg1b.List(oldRepo)
	if err != nil {
		return err
	}

	for _, cred := range creds {
		var newCred mg2b.Credential
		if mg2b.IdExist(newRepo, mg2e.Id(cred.ID().String())) {
			continue
		}

		switch cred := cred.(type) {
		case *mg1b.Login:
			newCred = &mg2b.Login{
				Login: cred.Login,
				CredentialBase: mg2b.NewCredentialBase(
					cred.Target(),
					cred.CreateTime(),
					cred.Salt(),
					cred.Metadata(),
				),
			}
		case *mg1b.LoginPassword:
			newCred = &mg2b.LoginPassword{
				Login:    cred.Login,
				Password: cred.Password,
				CredentialBase: mg2b.NewCredentialBase(
					cred.Target(),
					cred.CreateTime(),
					cred.Salt(),
					cred.Metadata(),
				),
			}
		case *mg1b.Token:
			newCred = &mg2b.Token{
				Value: cred.Value,
				CredentialBase: mg2b.NewCredentialBase(
					cred.Target(),
					cred.CreateTime(),
					cred.Salt(),
					cred.Metadata(),
				),
			}
		default:
			return errors.New("unknown credential encountered")
		}

		err = mg2b.Store(newRepo, newCred)
		if err != nil {
			return err
		}
	}

	return nil
}
