package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenSerial(t *testing.T) {
	original := NewToken("github", "value")
	loaded := testCredentialSerial(t, original)
	require.Equal(t, original.Value, loaded.(*Token).Value)
}
