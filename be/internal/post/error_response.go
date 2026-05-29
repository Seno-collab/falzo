package post

import (
	"errors"
	"net/http"

	"falzo-be/internal/share"
)

var (
	errInvalidJSONPayload = errors.New("invalid JSON payload")
	errInvalidPostIDParam = errors.New("invalid post id param")
	errInvalidPageParam   = errors.New("invalid page param")
	errInvalidLimitParam  = errors.New("invalid limit param")
)

func mapPostError(err error) share.ApiError {
	switch {
	case errors.Is(err, errInvalidJSONPayload):
		return share.ApiError{
			Status:  http.StatusBadRequest,
			Message: share.ValidationField,
			Code:    share.INVALID_FORMAT,
			Detail:  "Invalid JSON payload",
		}
	case errors.Is(err, errInvalidPostIDParam):
		return share.BadRequest("id", "id must be a valid positive integer")
	case errors.Is(err, errInvalidPageParam):
		return share.BadRequest("page", "page must be an integer")
	case errors.Is(err, errInvalidLimitParam):
		return share.BadRequest("limit", "limit must be an integer")
	case errors.Is(err, ErrUserIDRequired):
		return share.Required("user_id", "user_id is required")
	case errors.Is(err, ErrPostIDRequired):
		return share.Required("id", "post id is required")
	case errors.Is(err, ErrPageMustBePositive):
		return share.BadRequest("page", "page must be greater than 0")
	case errors.Is(err, ErrLimitMustBePositive):
		return share.BadRequest("limit", "limit must be greater than 0")
	case errors.Is(err, ErrLimitTooLarge):
		return share.BadRequest("limit", "limit must not exceed 50")
	case errors.Is(err, ErrInvalidCursor):
		return share.BadRequest("cursor", "cursor is invalid")
	case errors.Is(err, ErrInvalidFeed):
		return share.BadRequest("feed", "feed must be following when provided")
	case errors.Is(err, ErrLocationNameRequired):
		return share.Required("location_name", "location_name is required")
	case errors.Is(err, ErrLatitudeOutOfRange):
		return share.BadRequest("latitude", "latitude must be between -90 and 90")
	case errors.Is(err, ErrLongitudeOutOfRange):
		return share.BadRequest("longitude", "longitude must be between -180 and 180")
	case errors.Is(err, ErrImageURLRequired):
		return share.Required("image_url", "image_url is required")
	case errors.Is(err, ErrInvalidImageURL):
		return share.BadRequest("image_url", "image_url must be a valid URL")
	case errors.Is(err, ErrCaptionTooLong):
		return share.BadRequest("caption", "caption exceeds max length")
	case errors.Is(err, ErrLocationNameTooLong):
		return share.BadRequest("location_name", "location_name exceeds max length")
	case errors.Is(err, ErrCategoryNotFound):
		return share.BadRequest("category_id", "category does not exist")
	case errors.Is(err, ErrTooManyCategories):
		return share.BadRequest("category_ids", "category_ids must not contain more than 20 items")
	case errors.Is(err, ErrCommentRequired):
		return share.Required("content", "comment content is required")
	case errors.Is(err, ErrCommentTooLong):
		return share.BadRequest("content", "comment content exceeds max length")
	case errors.Is(err, ErrReplyCommentNotFound):
		return share.BadRequest("reply_to_comment_id", "reply comment does not exist on this post")
	case errors.Is(err, ErrCommentNotFound):
		return share.NotFound("Comment not found", "Requested comment does not exist")
	case errors.Is(err, ErrCommentUpdateForbidden):
		return share.ApiError{
			Status:  http.StatusForbidden,
			Message: "Forbidden",
			Code:    "FORBIDDEN",
			Detail:  "You can only edit your own comments",
		}
	case errors.Is(err, ErrCollectionIDRequired):
		return share.Required("collection_id", "collection id is required")
	case errors.Is(err, ErrCollectionNameRequired):
		return share.Required("name", "collection name is required")
	case errors.Is(err, ErrCollectionNameTooLong):
		return share.BadRequest("name", "collection name must not exceed 120 characters")
	case errors.Is(err, ErrCollectionNameTaken):
		return share.BadRequest("name", "collection name already exists")
	case errors.Is(err, ErrCollectionNotFound):
		return share.NotFound("Collection not found", "Requested saved collection does not exist")
	case errors.Is(err, ErrCollectionSlugRequired):
		return share.Required("share_slug", "collection share slug is required")
	case errors.Is(err, ErrPostUpdateForbidden):
		return share.ApiError{
			Status:  http.StatusForbidden,
			Message: "Forbidden",
			Code:    "FORBIDDEN",
			Detail:  "You can only edit your own posts",
		}
	case errors.Is(err, ErrPostModerationForbidden):
		return share.ApiError{
			Status:  http.StatusForbidden,
			Message: "Forbidden",
			Code:    "FORBIDDEN",
			Detail:  "You cannot moderate this content",
		}
	case errors.Is(err, ErrInvalidPostSort):
		return share.BadRequest("sort", "sort must be newest, popular, trending, or nearby")
	case errors.Is(err, ErrNearbyCoordinatesRequired):
		return share.BadRequest("lat", "nearby sort requires valid lat and lng query params")
	case errors.Is(err, ErrNearbyRadiusTooLarge):
		return share.BadRequest("radius_meters", "nearby radius must not exceed 1000 km")
	case errors.Is(err, ErrReportReasonRequired):
		return share.Required("reason", "reason is required")
	case errors.Is(err, ErrReportReasonTooLong):
		return share.BadRequest("reason", "reason must not exceed 500 characters")
	case errors.Is(err, ErrNotFound):
		return share.NotFound("Post not found", "Requested post does not exist")
	case errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Post service unavailable", "Post service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
