package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/post/application/command"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

type SavePostRequest struct {
	UserID uint64 `json:"user_id"`
}

func (h *Handler) SavePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		httpResponse.Error(w, http.StatusBadRequest, share.ValidationField, r, httpResponse.ErrorDetail{
			Code:    share.INVALID_FIELD,
			Field:   "id",
			Message: "id must be a valid positive integer",
		})
		return
	}

	var req SavePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse.Error(w, http.StatusBadRequest, share.ValidationField, r, httpResponse.ErrorDetail{
			Code:    share.INVALID_FORMAT,
			Message: "Invalid JSON payload",
		})
		return
	}

	if err := h.service.SavePost(r.Context(), command.SavePost{PostID: postID, UserID: req.UserID}); err != nil {
		writePostError(w, r, err, "save_post")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Post saved successfully", map[string]string{"message": "post saved"}, r)
}
