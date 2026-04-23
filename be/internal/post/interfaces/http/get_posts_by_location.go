package httpapi

import (
	"net/http"
	"strings"

	"falzo-be/internal/post/application/query"
	httpResponse "falzo-be/pkg/response"
)

func (h *Handler) GetPostsByLocation(w http.ResponseWriter, r *http.Request) {
	locationName := strings.TrimSpace(r.URL.Query().Get("location_name"))

	posts, err := h.service.GetPostsByLocation(r.Context(), query.GetPostsByLocation{LocationName: locationName})
	if err != nil {
		writePostError(w, r, err, "get_posts_by_location")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts by location fetched successfully", posts, r)
}
