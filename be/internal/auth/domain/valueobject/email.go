package valueobject

import (
	"errors"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email")

type Email string

func NewEmail(raw string) (Email, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" || !strings.Contains(value, "@") {
		return "", ErrInvalidEmail
	}

	return Email(value), nil
}

func (e Email) String() string {
	return string(e)
}
