package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c, err := New()
	require.NoError(t, err)
	require.NotNil(t, c)
}
