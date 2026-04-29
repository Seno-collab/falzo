package infra

import (
	"falzo-be/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

func NewPasswordHasher() PasswordHasher {
	return PasswordHasher{}
}

func (PasswordHasher) Hash(password auth.RawPassword) (auth.PasswordHash, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password.String()), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return auth.NewPasswordHash(string(hashed))
}

func (PasswordHasher) Compare(hash auth.PasswordHash, password auth.RawPassword) error {
	return bcrypt.CompareHashAndPassword([]byte(hash.String()), []byte(password.String()))
}

var _ auth.PasswordHasher = PasswordHasher{}
