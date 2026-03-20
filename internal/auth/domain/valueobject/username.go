package valueobject

import (
	"errors"
	"strings"
)

var ErrInvalidUsername = errors.New("invalid username")

type Username string

func NewUsername(raw string) (Username, error) {
	value := strings.TrimSpace(raw)
	if len(value) < 3 || len(value) > 50 {
		return "", ErrInvalidUsername
	}

	return Username(value), nil
}

func (u Username) String() string {
	return string(u)
}
