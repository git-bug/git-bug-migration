package auth

import (
	"crypto/sha256"
	"fmt"

	"github.com/MichaelMure/git-bug-migration/migration2/after/entity"
)

const (
	keyringKeyTokenValue = "value"
)

var _ Credential = &Token{}

// Token holds an API access token data
type Token struct {
	*CredentialBase
	Value string
}

// NewToken instantiate a new token
func NewToken(target, value string) *Token {
	return &Token{
		CredentialBase: newCredentialBase(target),
		Value:          value,
	}
}

func NewTokenFromConfig(conf map[string]string) (*Token, error) {
	base, err := newCredentialBaseFromData(conf)
	if err != nil {
		return nil, err
	}

	return &Token{
		CredentialBase: base,
		Value:          conf[keyringKeyTokenValue],
	}, nil
}

func (t *Token) ID() entity.Id {
	h := sha256.New()
	_, _ = h.Write(t.SaltT)
	_, _ = h.Write([]byte(t.TargetT))
	_, _ = h.Write([]byte(t.Value))
	return entity.Id(fmt.Sprintf("%x", h.Sum(nil)))
}

func (t *Token) Kind() CredentialKind {
	return KindToken
}

// Validate ensure token important fields are valid
func (t *Token) Validate() error {
	err := t.CredentialBase.validate()
	if err != nil {
		return err
	}
	if t.Value == "" {
		return fmt.Errorf("missing value")
	}
	return nil
}

func (t *Token) toConfig() map[string]string {
	return map[string]string{
		keyringKeyTokenValue: t.Value,
	}
}
