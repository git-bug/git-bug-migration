package auth

import (
	"crypto/sha256"
	"fmt"

	"github.com/MichaelMure/git-bug-migration/migration2/after/entity"
)

const (
	keyringKeyLoginLogin = "login"
)

var _ Credential = &Login{}

type Login struct {
	*CredentialBase
	Login string
}

func NewLogin(target, login string) *Login {
	return &Login{
		CredentialBase: newCredentialBase(target),
		Login:          login,
	}
}

func NewLoginFromConfig(conf map[string]string) (*Login, error) {
	base, err := newCredentialBaseFromData(conf)
	if err != nil {
		return nil, err
	}

	return &Login{
		CredentialBase: base,
		Login:          conf[keyringKeyLoginLogin],
	}, nil
}

func (lp *Login) ID() entity.Id {
	h := sha256.New()
	_, _ = h.Write(lp.SaltT)
	_, _ = h.Write([]byte(lp.TargetT))
	_, _ = h.Write([]byte(lp.Login))
	return entity.Id(fmt.Sprintf("%x", h.Sum(nil)))
}

func (lp *Login) Kind() CredentialKind {
	return KindLogin
}

func (lp *Login) Validate() error {
	err := lp.CredentialBase.validate()
	if err != nil {
		return err
	}
	if lp.Login == "" {
		return fmt.Errorf("missing login")
	}
	return nil
}

func (lp *Login) toConfig() map[string]string {
	return map[string]string{
		keyringKeyLoginLogin: lp.Login,
	}
}
