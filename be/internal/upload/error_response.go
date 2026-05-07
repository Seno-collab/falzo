package upload

import (
	"errors"
	"net/http"

	"falzo-be/internal/share"
)

var errInvalidMultipartPayload = errors.New("invalid multipart payload")

func mapUploadError(err error) share.ApiError {
	switch {
	case errors.Is(err, errInvalidMultipartPayload):
		return share.ApiError{
			Status:  http.StatusBadRequest,
			Message: share.ValidationField,
			Code:    share.INVALID_FORMAT,
			Detail:  "Invalid multipart payload",
		}
	case errors.Is(err, ErrFileRequired):
		return share.Required("file", "file is required")
	case errors.Is(err, ErrOwnerIDRequired):
		return share.Required("owner_id", "owner_id is required")
	case errors.Is(err, ErrInvalidFileSize):
		return share.BadRequest("file", "file size is invalid")
	case errors.Is(err, ErrFileSizeTooLarge):
		return share.BadRequest("file", "file size exceeds the maximum allowed limit")
	case errors.Is(err, ErrInvalidMimeType):
		return share.BadRequest("file", "file mime type is invalid")
	case errors.Is(err, ErrInvalidImageURL):
		return share.BadRequest("url", "image URL is invalid")
	case errors.Is(err, ErrImageNotFound):
		return share.NotFound("Image not found", "Requested image does not exist")
	case errors.Is(err, ErrImageIDRequired):
		return share.BadRequest("id", "image id is required")
	case errors.Is(err, ErrStorageFailed), errors.Is(err, ErrMissingRepository), errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Image upload unavailable", "Image upload is temporarily unavailable")
	default:
		return share.Internal()
	}
}
