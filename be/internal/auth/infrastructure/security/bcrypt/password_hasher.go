package bcrypt

import (
	domainservice "falzo-be/internal/auth/domain/service"
	"falzo-be/internal/auth/domain/value_object"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

func NewPasswordHasher() domainservice.PasswordHasher {
	return PasswordHasher{}
}

func (PasswordHasher) Hash(password value_object.RawPassword) (value_object.PasswordHash, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password.String()), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return value_object.NewPasswordHash(string(hashed))
}

func (PasswordHasher) Compare(hash value_object.PasswordHash, password value_object.RawPassword) error {
	return bcrypt.CompareHashAndPassword([]byte(hash.String()), []byte(password.String()))
}
