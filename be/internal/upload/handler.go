package upload

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

const (
	defaultMultipartMemory     = 10 << 20
	multipartOverheadAllowance = 1 << 20
	defaultUploadMaxBodyBytes  = defaultMaxImageSize + multipartOverheadAllowance
)

type handlerService interface {
	UploadImage(ctx context.Context, input UploadImageInput) (UploadImageResult, error)
	UpdateImage(ctx context.Context, input UpdateImageInput) (UpdateImageResult, error)
}

type Handler struct {
	service     handlerService
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	}
	protectedMiddlewares []func(http.Handler) http.Handler
	maxBodyBytes         int64
}

type HandlerOption func(*Handler)

func WithProtectedMiddlewares(middlewares ...func(http.Handler) http.Handler) HandlerOption {
	return func(h *Handler) {
		h.protectedMiddlewares = append(h.protectedMiddlewares, middlewares...)
	}
}

func WithMaxBodyBytes(maxBodyBytes int64) HandlerOption {
	return func(h *Handler) {
		if maxBodyBytes > 0 {
			h.maxBodyBytes = maxBodyBytes
		}
	}
}

func NewHandler(
	service handlerService,
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	},
	options ...HandlerOption,
) *Handler {
	h := &Handler{
		service:      service,
		authService:  authService,
		maxBodyBytes: defaultUploadMaxBodyBytes,
	}
	for _, option := range options {
		option(h)
	}
	return h
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		for _, middleware := range h.protectedMiddlewares {
			protected.Use(middleware)
		}
		protected.Post("/images/upload", h.UploadImage)
		protected.Put("/images/{id}", h.UpdateImage)
	})
	return r
}

type UploadImageRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if !h.parseMultipartForm(w, r, "upload_image") {
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

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrOwnerIDRequired, "upload_image", mapUploadError)
		return
	}

	result, err := h.service.UploadImage(r.Context(), UploadImageInput{
		File:     data,
		FileName: header.Filename,
		MimeType: detectImageMimeType(data, header),
		Size:     header.Size,
		OwnerID:  strconv.FormatUint(principal.UserID, 10),
	})
	if err != nil {
		share.WriteError(w, r, err, "upload_image", mapUploadError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Image uploaded successfully", result, r)
}

func (h *Handler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	imageID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || imageID <= 0 {
		share.WriteError(w, r, ErrImageIDRequired, "update_image", mapUploadError)
		return
	}
	if !h.parseMultipartForm(w, r, "update_image") {
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		share.WriteError(w, r, ErrFileRequired, "update_image", mapUploadError)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		share.WriteError(w, r, ErrStorageFailed, "update_image", mapUploadError)
		return
	}

	result, err := h.service.UpdateImage(r.Context(), UpdateImageInput{
		ImageID:  imageID,
		File:     data,
		FileName: header.Filename,
		MimeType: detectImageMimeType(data, header),
		Size:     header.Size,
	})
	if err != nil {
		share.WriteError(w, r, err, "update_image", mapUploadError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Image updated successfully", result, r)
}

func (h *Handler) parseMultipartForm(w http.ResponseWriter, r *http.Request, operation string) bool {
	if h.maxBodyBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	}

	if err := r.ParseMultipartForm(defaultMultipartMemory); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			share.WriteError(w, r, ErrFileSizeTooLarge, operation, mapUploadError)
			return false
		}

		share.WriteError(w, r, errInvalidMultipartPayload, operation, mapUploadError)
		return false
	}

	return true
}

func detectImageMimeType(data []byte, header *multipart.FileHeader) string {
	headerType := ""
	if header != nil {
		headerType = header.Header.Get("Content-Type")
	}
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	if len(data) > 0 {
		detected := http.DetectContentType(data)
		if detected != "application/octet-stream" {
			return detected
		}
	}
	return headerType
}
