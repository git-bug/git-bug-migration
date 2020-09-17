package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MichaelMure/git-bug-migration/migration12/before/entity"
	"github.com/MichaelMure/git-bug-migration/migration12/before/repository"
)

func TestCredential(t *testing.T) {
	repo := repository.NewMockRepoForTest()

	storeToken := func(val string, target string) *Token {
		token := NewToken(target, val)
		err := Store(repo, token)
		require.NoError(t, err)
		return token
	}

	token := storeToken("foobar", "github")

	// Store + Load
	err := Store(repo, token)
	require.NoError(t, err)

	token2, err := LoadWithId(repo, token.ID())
	require.NoError(t, err)
	require.Equal(t, token.CreateTimeT.Unix(), token2.CreateTime().Unix())
	token.CreateTimeT = token2.CreateTime()
	require.Equal(t, token, token2)

	prefix := string(token.ID())[:10]

	// LoadWithPrefix
	token3, err := LoadWithPrefix(repo, prefix)
	require.NoError(t, err)
	require.Equal(t, token.CreateTimeT.Unix(), token3.CreateTime().Unix())
	token.CreateTimeT = token3.CreateTime()
	require.Equal(t, token, token3)

	token4 := storeToken("foo", "gitlab")
	token5 := storeToken("bar", "github")

	// List + options
	creds, err := List(repo, WithTarget("github"))
	require.NoError(t, err)
	sameIds(t, creds, []Credential{token, token5})

	creds, err = List(repo, WithTarget("gitlab"))
	require.NoError(t, err)
	sameIds(t, creds, []Credential{token4})

	creds, err = List(repo, WithKind(KindToken))
	require.NoError(t, err)
	sameIds(t, creds, []Credential{token, token4, token5})

	creds, err = List(repo, WithKind(KindLoginPassword))
	require.NoError(t, err)
	sameIds(t, creds, []Credential{})

	// Metadata

	token4.SetMetadata("key", "value")
	err = Store(repo, token4)
	require.NoError(t, err)

	creds, err = List(repo, WithMeta("key", "value"))
	require.NoError(t, err)
	sameIds(t, creds, []Credential{token4})

	// Exist
	exist := IdExist(repo, token.ID())
	require.True(t, exist)

	exist = PrefixExist(repo, prefix)
	require.True(t, exist)

	// Remove
	err = Remove(repo, token.ID())
	require.NoError(t, err)

	creds, err = List(repo)
	require.NoError(t, err)
	sameIds(t, creds, []Credential{token4, token5})
}

func sameIds(t *testing.T, a []Credential, b []Credential) {
	t.Helper()

	ids := func(creds []Credential) []entity.Id {
		result := make([]entity.Id, len(creds))
		for i, cred := range creds {
			result[i] = cred.ID()
		}
		return result
	}

	require.ElementsMatch(t, ids(a), ids(b))
}

func testCredentialSerial(t *testing.T, original Credential) Credential {
	repo := repository.NewMockRepoForTest()

	original.SetMetadata("test", "value")

	require.NotEmpty(t, original.ID().String())
	require.NotEmpty(t, original.Salt())
	require.NoError(t, Store(repo, original))

	loaded, err := LoadWithId(repo, original.ID())
	require.NoError(t, err)

	require.Equal(t, original.ID(), loaded.ID())
	require.Equal(t, original.Kind(), loaded.Kind())
	require.Equal(t, original.Target(), loaded.Target())
	require.Equal(t, original.CreateTime().Unix(), loaded.CreateTime().Unix())
	require.Equal(t, original.Salt(), loaded.Salt())
	require.Equal(t, original.Metadata(), loaded.Metadata())

	return loaded
}
