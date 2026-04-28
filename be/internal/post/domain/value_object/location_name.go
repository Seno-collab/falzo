package value_object

import (
	"errors"
	"strings"
)

const maxLocationNameLength = 255

var ErrLocationNameTooLong = errors.New("location name exceeds max length")

type LocationName struct {
	value string
}

func NewLocationName(raw string) (LocationName, error) {
	value := strings.TrimSpace(raw)
	if len(value) > maxLocationNameLength {
		return LocationName{}, ErrLocationNameTooLong
	}

	return LocationName{value: value}, nil
}

func (l LocationName) String() string {
	return l.value
}
