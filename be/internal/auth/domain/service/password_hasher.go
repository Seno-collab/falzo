package service

import "falzo-be/internal/auth/domain/valueobject"

type PasswordHasher interface {
	Hash(password valueobject.RawPassword) (valueobject.PasswordHash, error)
	Compare(hash valueobject.PasswordHash, password valueobject.RawPassword) error
}
