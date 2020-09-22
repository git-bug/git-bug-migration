package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginPasswordSerial(t *testing.T) {
	original := NewLoginPassword("github", "jean", "jacques")
	loaded := testCredentialSerial(t, original)
	require.Equal(t, original.Login, loaded.(*LoginPassword).Login)
	require.Equal(t, original.Password, loaded.(*LoginPassword).Password)
}
