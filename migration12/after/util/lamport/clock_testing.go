package lamport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testClock(t *testing.T, c Clock) {
	require.Equal(t, Time(1), c.Time())

	val, err := c.Increment()
	require.NoError(t, err)
	require.Equal(t, Time(1), val)
	require.Equal(t, Time(2), c.Time())

	err = c.Witness(41)
	require.NoError(t, err)
	require.Equal(t, Time(42), c.Time())

	err = c.Witness(41)
	require.NoError(t, err)
	require.Equal(t, Time(42), c.Time())

	err = c.Witness(30)
	require.NoError(t, err)
	require.Equal(t, Time(42), c.Time())
}
