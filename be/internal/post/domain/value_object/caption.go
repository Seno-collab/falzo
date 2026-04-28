package value_object

import (
	"errors"
	"strings"
)

const maxCaptionLength = 2000

var ErrCaptionTooLong = errors.New("caption exceeds max length")

type Caption struct {
	value string
}

func NewCaption(raw string) (Caption, error) {
	value := strings.TrimSpace(raw)
	if len(value) > maxCaptionLength {
		return Caption{}, ErrCaptionTooLong
	}

	return Caption{value: value}, nil
}

func (c Caption) String() string {
	return c.value
}
