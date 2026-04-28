package value_object

import (
	"errors"
	"unicode"
)

var ErrInvalidPassword = errors.New("invalid password")

type RawPassword string

func NewRawPassword(raw string) (RawPassword, error) {
	if len(raw) < 8 {
		return "", ErrInvalidPassword
	}

	hasLetter := false
	hasDigit := false
	for _, r := range raw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "", ErrInvalidPassword
	}

	return RawPassword(raw), nil
}

func (p RawPassword) String() string {
	return string(p)
}
