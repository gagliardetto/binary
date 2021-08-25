package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSighash(t *testing.T) {
	value := "hello"
	got := Sighash(SIGHASH_GLOBAL_NAMESPACE, value)

	require.NotEmpty(t, got)

	expected := []byte{149, 118, 59, 220, 196, 127, 161, 179}
	require.Equal(t,
		expected,
		got,
	)
}
