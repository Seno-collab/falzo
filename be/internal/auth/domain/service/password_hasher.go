package service

import "falzo-be/internal/auth/domain/value_object"

type PasswordHasher interface {
	Hash(password value_object.RawPassword) (value_object.PasswordHash, error)
	Compare(hash value_object.PasswordHash, password value_object.RawPassword) error
}
