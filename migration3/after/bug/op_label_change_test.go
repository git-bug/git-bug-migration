package bug

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/MichaelMure/git-bug-migration/migration3/after/identity"
	"github.com/MichaelMure/git-bug-migration/migration3/after/repository"
)

func TestLabelChangeSerialize(t *testing.T) {
	repo := repository.NewMockRepoForTest()

	rene, err := identity.NewIdentity(repo, "René Descartes", "rene@descartes.fr")
	require.NoError(t, err)
	err = rene.Commit(repo)
	require.NoError(t, err)

	unix := time.Now().Unix()
	before := NewLabelChangeOperation(rene, unix, []Label{"added"}, []Label{"removed"})

	data, err := json.Marshal(before)
	require.NoError(t, err)

	var after LabelChangeOperation
	err = json.Unmarshal(data, &after)
	require.NoError(t, err)

	// enforce creating the ID
	before.Id()

	// Replace the identity stub with the real thing
	require.Equal(t, rene.Id(), after.base().Author.Id())
	after.Author = rene

	require.Equal(t, before, &after)
}
