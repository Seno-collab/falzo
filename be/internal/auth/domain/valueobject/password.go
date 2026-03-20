package valueobject

import "errors"

var ErrInvalidPassword = errors.New("invalid password")

type RawPassword string

func NewRawPassword(raw string) (RawPassword, error) {
	if len(raw) < 6 {
		return "", ErrInvalidPassword
	}

	return RawPassword(raw), nil
}

func (p RawPassword) String() string {
	return string(p)
}
