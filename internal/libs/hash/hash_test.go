package hash_test

import (
	"testing"

	"github.com/Meat-Hook/back-template/internal/libs/hash"
	"github.com/stretchr/testify/assert"
)

var (
	pass = "pass"
)

func TestHasher_Smoke(t *testing.T) {
	t.Parallel()

	passwords := hash.New()
	hashPass, err := passwords.Hashing(pass)
	assert.NoError(t, err)
	compare := passwords.Compare(hashPass, []byte(pass))
	assert.Equal(t, true, compare)
}
