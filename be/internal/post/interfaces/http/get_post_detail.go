package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/post/application/query"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetPostDetail(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		httpResponse.Error(w, http.StatusBadRequest, share.ValidationField, r, httpResponse.ErrorDetail{
			Code:    share.INVALID_FIELD,
			Field:   "id",
			Message: "id must be a valid positive integer",
		})
		return
	}

	post, err := h.service.GetPostDetail(r.Context(), query.GetPostDetail{PostID: postID})
	if err != nil {
		writePostError(w, r, err, "get_post_detail")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Post detail fetched successfully", post, r)
}
