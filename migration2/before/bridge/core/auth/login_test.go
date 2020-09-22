package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginSerial(t *testing.T) {
	original := NewLogin("github", "jean")
	loaded := testCredentialSerial(t, original)
	require.Equal(t, original.Login, loaded.(*Login).Login)
}
