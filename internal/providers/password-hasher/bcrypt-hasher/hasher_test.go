package bcrypt_hasher

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	hasher := New(10)

	password := "mysecurepassword"
	hashedPassword, err := hasher.GenerateFromPassword(password)

	assert.NoError(t, err)
	assert.NotEqual(t, password, hashedPassword)
	assert.NotEmpty(t, hashedPassword)
}

func TestCompare(t *testing.T) {
	hasher := New(10)

	password := "mysecurepassword"
	hashedPassword, err := hasher.GenerateFromPassword(password)
	assert.NoError(t, err)

	err = hasher.CompareHashAndPassword(hashedPassword, password)
	assert.NoError(t, err)

	wrongPassword := "wrongpassword"
	err = hasher.CompareHashAndPassword(hashedPassword, wrongPassword)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCompare))
}
