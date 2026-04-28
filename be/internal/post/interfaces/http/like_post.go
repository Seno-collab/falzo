package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/post/application/command"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

type LikePostRequest struct {
	UserID uint64 `json:"user_id"`
}

func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		httpResponse.Error(w, http.StatusBadRequest, "ValidationField", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FIELD",
			Field:   "id",
			Message: "id must be a valid positive integer",
		})
		return
	}

	var req LikePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "ValidationField", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if err := h.service.LikePost(r.Context(), command.LikePost{PostID: postID, UserID: req.UserID}); err != nil {
		writePostError(w, r, err, "like_post")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Post liked successfully", map[string]string{"message": "post liked"}, r)
}
