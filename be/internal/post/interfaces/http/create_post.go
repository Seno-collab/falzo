package httpapi

import (
	"encoding/json"
	"net/http"

	"falzo-be/internal/post/application/command"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"
)

type CreatePostRequest struct {
	UserID       uint64  `json:"user_id"`
	ImageURL     string  `json:"image_url"`
	Caption      string  `json:"caption"`
	LocationName string  `json:"location_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse.Error(w, http.StatusBadRequest, share.ValidationField, r, httpResponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	post, err := h.service.CreatePost(r.Context(), command.CreatePost{
		UserID:       req.UserID,
		ImageURL:     req.ImageURL,
		Caption:      req.Caption,
		LocationName: req.LocationName,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
	})
	if err != nil {
		writePostError(w, r, err, "create_post")
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Post created successfully", post, r)
}
