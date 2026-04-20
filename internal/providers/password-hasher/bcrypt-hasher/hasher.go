package bcrypt_hasher

import (
	"errors"

	"github.com/tiptop-co/backend/pkg/errwrap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrHash    = errors.New("hashing error")
	ErrCompare = errors.New("compare error")
)

type BcryptPasswordHasher struct {
	hashCost int
}

func New(hashCost int) *BcryptPasswordHasher {
	return &BcryptPasswordHasher{
		hashCost: hashCost,
	}
}

func (b *BcryptPasswordHasher) GenerateFromPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), b.hashCost)

	return string(hashed), errwrap.Wrap(ErrHash, err)
}

func (BcryptPasswordHasher) CompareHashAndPassword(hashed, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))

	return errwrap.Wrap(ErrCompare, err)
}
