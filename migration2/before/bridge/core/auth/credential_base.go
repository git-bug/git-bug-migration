package auth

import (
	"fmt"
	"time"

	"github.com/MichaelMure/git-bug-migration/migration2/before/repository"
)

type CredentialBase struct {
	TargetT     string            `json:"target"`
	CreateTimeT time.Time         `json:"create_time"`
	SaltT       []byte            `json:"salt"`
	MetaT       map[string]string `json:"meta"`
}

func newCredentialBase(target string) *CredentialBase {
	return &CredentialBase{
		TargetT:     target,
		CreateTimeT: time.Now(),
		SaltT:       makeSalt(),
	}
}

func newCredentialBaseFromConfig(conf map[string]string) (*CredentialBase, error) {
	base := &CredentialBase{
		TargetT: conf[configKeyTarget],
		MetaT:   metaFromConfig(conf),
	}

	if createTime, ok := conf[configKeyCreateTime]; ok {
		t, err := repository.ParseTimestamp(createTime)
		if err != nil {
			return nil, err
		}
		base.CreateTimeT = t
	} else {
		return nil, fmt.Errorf("missing create time")
	}

	salt, err := saltFromConfig(conf)
	if err != nil {
		return nil, err
	}
	base.SaltT = salt

	return base, nil
}

func (cb *CredentialBase) Target() string {
	return cb.TargetT
}

func (cb *CredentialBase) CreateTime() time.Time {
	return cb.CreateTimeT
}

func (cb *CredentialBase) Salt() []byte {
	return cb.SaltT
}

func (cb *CredentialBase) validate() error {
	if cb.TargetT == "" {
		return fmt.Errorf("missing target")
	}
	if cb.CreateTimeT.IsZero() || cb.CreateTimeT.Equal(time.Time{}) {
		return fmt.Errorf("missing creation time")
	}
	return nil
}

func (cb *CredentialBase) Metadata() map[string]string {
	return cb.MetaT
}

func (cb *CredentialBase) GetMetadata(key string) (string, bool) {
	val, ok := cb.MetaT[key]
	return val, ok
}

func (cb *CredentialBase) SetMetadata(key string, value string) {
	if cb.MetaT == nil {
		cb.MetaT = make(map[string]string)
	}
	cb.MetaT[key] = value
}
