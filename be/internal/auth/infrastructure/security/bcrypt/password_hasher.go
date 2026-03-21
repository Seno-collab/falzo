package bcrypt

import (
	domainservice "falzo-be/internal/auth/domain/service"
	"falzo-be/internal/auth/domain/valueobject"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

func NewPasswordHasher() domainservice.PasswordHasher {
	return PasswordHasher{}
}

func (PasswordHasher) Hash(password valueobject.RawPassword) (valueobject.PasswordHash, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password.String()), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return valueobject.NewPasswordHash(string(hashed))
}

func (PasswordHasher) Compare(hash valueobject.PasswordHash, password valueobject.RawPassword) error {
	return bcrypt.CompareHashAndPassword([]byte(hash.String()), []byte(password.String()))
}
