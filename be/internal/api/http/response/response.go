package response

import (
	"be/internal/shared/apperror"
	"encoding/json"
	"net/http"
)

type APIResponse[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusOK, APIResponse[T]{
		Code:    "SUCCESS",
		Message: "Success",
		Data:    data,
	})
}

func Created[T any](w http.ResponseWriter, data T) {
	JSON(w, http.StatusCreated, APIResponse[T]{
		Code:    "SUCCESS",
		Message: "Created successfully",
		Data:    data,
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, err error) {
	appErr := apperror.FromError(err)

	JSON(w, appErr.HTTPStatus, ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Message,
		Details: appErr.Details,
	})
}
