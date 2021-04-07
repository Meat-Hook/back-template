package hash_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/libs/hash"
	"github.com/stretchr/testify/require"
)

var pass = "pass"

func TestHasher_Smoke(t *testing.T) {
	t.Parallel()

	passwords := hash.New()
	assert := require.New(t)
	hashPass, err := passwords.Hashing(pass)
	assert.NoError(err)
	compare := passwords.Compare(hashPass, []byte(pass))
	assert.Equal(true, compare)
}
