package value_object

import "errors"

var ErrInvalidPasswordHash = errors.New("invalid password hash")

type PasswordHash string

func NewPasswordHash(raw string) (PasswordHash, error) {
	if raw == "" {
		return "", ErrInvalidPasswordHash
	}

	return PasswordHash(raw), nil
}

func (p PasswordHash) String() string {
	return string(p)
}
