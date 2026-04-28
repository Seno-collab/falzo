package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/post/application/query"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"
)

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 12

	pageRaw := strings.TrimSpace(r.URL.Query().Get("page"))
	if pageRaw != "" {
		parsedPage, err := strconv.Atoi(pageRaw)
		if err != nil {
			httpResponse.Error(w, http.StatusBadRequest, share.ValidationField, r, httpResponse.ErrorDetail{
				Code:    share.INVALID_FIELD,
				Field:   "page",
				Message: "page must be an integer",
			})
			return
		}
		page = parsedPage
	}

	limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if limitRaw != "" {
		parsedLimit, err := strconv.Atoi(limitRaw)
		if err != nil {
			httpResponse.Error(w, http.StatusBadRequest, share.ValidationField, r, httpResponse.ErrorDetail{
				Code:    share.INVALID_FIELD,
				Field:   "limit",
				Message: "limit must be an integer",
			})
			return
		}
		limit = parsedLimit
	}

	posts, err := h.service.GetPosts(r.Context(), query.GetPosts{Page: page, Limit: limit})
	if err != nil {
		writePostError(w, r, err, "get_posts")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts fetched successfully", posts, r)
}
