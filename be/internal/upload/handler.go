package upload

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

var errInvalidMultipartPayload = errors.New("invalid multipart payload")

type handlerService interface {
	UploadImage(ctx context.Context, input UploadImageInput) (UploadImageResult, error)
}

type Handler struct {
	service     handlerService
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	}
}

func NewHandler(
	service handlerService,
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	},
) *Handler {
	return &Handler{service: service, authService: authService}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Post("/images", h.UploadImage)
	})
	return r
}

type UploadImageRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		share.WriteError(w, r, errInvalidMultipartPayload, "upload_image", mapUploadError)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		share.WriteError(w, r, ErrFileRequired, "upload_image", mapUploadError)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		share.WriteError(w, r, ErrStorageFailed, "upload_image", mapUploadError)
		return
	}

	ownerID := r.FormValue("owner_id")
	result, err := h.service.UploadImage(r.Context(), UploadImageInput{
		File:     data,
		FileName: header.Filename,
		MimeType: header.Header.Get("Content-Type"),
		Size:     header.Size,
		OwnerID:  ownerID,
	})
	if err != nil {
		share.WriteError(w, r, err, "upload_image", mapUploadError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Image uploaded successfully", result, r)
}

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
	case errors.Is(err, ErrStorageFailed), errors.Is(err, ErrMissingRepository):
		return share.ServiceUnavailable("Image upload unavailable", "Image upload is temporarily unavailable")
	default:
		return share.Internal()
	}
}
