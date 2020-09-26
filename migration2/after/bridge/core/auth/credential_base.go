package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/MichaelMure/git-bug-migration/migration2/after/repository"
)

type CredentialBase struct {
	TargetT     string            `json:"target"`
	CreateTimeT time.Time         `json:"create_time"`
	SaltT       []byte            `json:"salt"`
	MetaT       map[string]string `json:"meta"`
}

func NewCredentialBase(target string, createTime time.Time, salt []byte, meta map[string]string) *CredentialBase {
	return &CredentialBase{TargetT: target, CreateTimeT: createTime, SaltT: salt, MetaT: meta}
}

func newCredentialBase(target string) *CredentialBase {
	return &CredentialBase{
		TargetT:     target,
		CreateTimeT: time.Now(),
		SaltT:       makeSalt(),
	}
}

func makeSalt() []byte {
	result := make([]byte, 16)
	_, err := rand.Read(result)
	if err != nil {
		panic(err)
	}
	return result
}

func newCredentialBaseFromData(data map[string]string) (*CredentialBase, error) {
	base := &CredentialBase{
		TargetT: data[keyringKeyTarget],
		MetaT:   metaFromData(data),
	}

	if createTime, ok := data[keyringKeyCreateTime]; ok {
		t, err := repository.ParseTimestamp(createTime)
		if err != nil {
			return nil, err
		}
		base.CreateTimeT = t
	} else {
		return nil, fmt.Errorf("missing create time")
	}

	salt, err := saltFromData(data)
	if err != nil {
		return nil, err
	}
	base.SaltT = salt

	return base, nil
}

func metaFromData(data map[string]string) map[string]string {
	result := make(map[string]string)
	for key, val := range data {
		if strings.HasPrefix(key, keyringKeyPrefixMeta) {
			key = strings.TrimPrefix(key, keyringKeyPrefixMeta)
			result[key] = val
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func saltFromData(data map[string]string) ([]byte, error) {
	val, ok := data[keyringKeySalt]
	if !ok {
		return nil, fmt.Errorf("no credential SaltT found")
	}
	return base64.StdEncoding.DecodeString(val)
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
