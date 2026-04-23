package valueobject

import (
	"errors"
	"net/url"
	"strings"
)

var ErrImageURLRequired = errors.New("image url is required")
var ErrInvalidImageURL = errors.New("invalid image url")

type ImageURL struct {
	value string
}

func NewImageURL(raw string) (ImageURL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ImageURL{}, ErrImageURLRequired
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return ImageURL{}, ErrInvalidImageURL
	}

	return ImageURL{value: value}, nil
}

func (i ImageURL) String() string {
	return i.value
}
