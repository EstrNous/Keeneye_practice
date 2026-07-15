package tokenutil_test

import (
	"testing"

	"keeneye_practice/app/internal/tokenutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashDeterministic(t *testing.T) {
	h1 := tokenutil.Hash("abc")
	h2 := tokenutil.Hash("abc")
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, "abc", h1)
}

func TestNewTokenUnique(t *testing.T) {
	raw1, hash1, err := tokenutil.NewToken()
	require.NoError(t, err)
	raw2, hash2, err := tokenutil.NewToken()
	require.NoError(t, err)
	assert.NotEqual(t, raw1, raw2)
	assert.NotEqual(t, hash1, hash2)
	assert.Equal(t, tokenutil.Hash(raw1), hash1)
}
