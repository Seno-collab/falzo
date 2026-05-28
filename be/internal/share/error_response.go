package share

import (
	"errors"
	httpResponse "falzo-be/pkg/response"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/rs/zerolog/log"
)

type ApiError struct {
	Status  int
	Message string
	Code    string
	Field   string
	Detail  string
	LogErr  bool
}

func WriteError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	operation string,
	mapper func(error) ApiError,
) {
	mapped := mapper(err)

	if mapped.LogErr {
		source := errorLogSource(2)
		logErr := err

		entry := log.Error().
			Err(err).
			Str("module", source.Module).
			Str("operation", operation).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("source_file", source.File).
			Int("source_line", source.Line).
			Str("source_func", source.Function).
			Int("status", mapped.Status).
			Str("api_code", mapped.Code).
			Str("api_detail", mapped.Detail)

		var appErr *AppError
		if errors.As(err, &appErr) {
			if appErr.Code != "" {
				entry = entry.Str("app_code", appErr.Code)
			}
			if appErr.Operation != "" {
				entry = entry.Str("app_operation", appErr.Operation)
			}
			if len(appErr.Metadata) > 0 {
				entry = entry.Interface("app_metadata", appErr.Metadata)
			}
			if appErr.Internal != nil {
				logErr = appErr.Internal
				entry = entry.AnErr("internal_error", appErr.Internal)
			}
		}

		chain := errorChain(logErr)
		if len(chain) > 0 {
			entry = entry.Strs("error_chain", chain).Str("root_error", chain[len(chain)-1])
		}

		entry.Msg("request failed")
	}

	httpResponse.Error(w, mapped.Status, mapped.Message, r, httpResponse.ErrorDetail{
		Code:    mapped.Code,
		Field:   mapped.Field,
		Message: mapped.Detail,
	})

}

type logSource struct {
	Module   string
	File     string
	Line     int
	Function string
}

func errorLogSource(skip int) logSource {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return logSource{}
	}

	fn := runtime.FuncForPC(pc)
	function := ""
	if fn != nil {
		function = fn.Name()
	}

	return logSource{
		Module:   moduleFromFunction(function, file),
		File:     trimSourceFile(file),
		Line:     line,
		Function: function,
	}
}

func moduleFromFunction(function, file string) string {
	pkg := functionPackage(function)
	if pkg == "" {
		pkg = filepath.Dir(file)
	}

	if module := packageModule(pkg, "/internal/"); module != "" {
		return module
	}
	if module := packageModule(pkg, "/pkg/"); module != "" {
		return module
	}

	return filepath.Base(pkg)
}

func functionPackage(function string) string {
	if function == "" {
		return ""
	}

	slash := strings.LastIndex(function, "/")
	if slash >= 0 {
		if idx := strings.Index(function[slash+1:], "."); idx >= 0 {
			return function[:slash+1+idx]
		}
	}

	if idx := strings.Index(function, "."); idx >= 0 {
		return function[:idx]
	}

	return function
}

func packageModule(pkg, marker string) string {
	idx := strings.Index(pkg, marker)
	if idx < 0 {
		return ""
	}

	module := pkg[idx+len(marker):]
	if marker == "/internal/" {
		return firstPathSegment(module)
	}
	return module
}

func firstPathSegment(path string) string {
	if idx := strings.Index(path, "/"); idx >= 0 {
		return path[:idx]
	}
	return path
}

func trimSourceFile(file string) string {
	for _, marker := range []string{"/internal/", "/pkg/"} {
		if idx := strings.Index(file, marker); idx >= 0 {
			return strings.TrimPrefix(file[idx:], "/")
		}
	}

	return filepath.Base(file)
}

func errorChain(err error) []string {
	if err == nil {
		return nil
	}

	chain := make([]string, 0, 4)
	seen := map[string]struct{}{}
	collectErrorChain(err, &chain, seen, 8)
	return chain
}

func collectErrorChain(err error, chain *[]string, seen map[string]struct{}, remaining int) {
	if err == nil || remaining <= 0 {
		return
	}

	message := err.Error()
	if message != "" {
		if _, ok := seen[message]; !ok {
			seen[message] = struct{}{}
			*chain = append(*chain, message)
		}
	}

	var multi interface {
		Unwrap() []error
	}
	if errors.As(err, &multi) {
		for _, inner := range multi.Unwrap() {
			collectErrorChain(inner, chain, seen, remaining-1)
		}
		return
	}

	collectErrorChain(errors.Unwrap(err), chain, seen, remaining-1)
}

func BadRequest(field, detail string) ApiError {
	return ApiError{
		Status:  http.StatusBadRequest,
		Message: ValidationField,
		Code:    INVALID_FIELD,
		Field:   field,
		Detail:  detail,
	}
}

func Required(field, detail string) ApiError {
	return ApiError{
		Status:  http.StatusBadRequest,
		Message: ValidationField,
		Code:    REQUIRED_FIELD,
		Field:   field,
		Detail:  detail,
	}
}

func NotFound(msg, detail string) ApiError {
	return ApiError{
		Status:  http.StatusNotFound,
		Message: msg,
		Code:    NOT_FOUND,
		Detail:  detail,
	}
}

func Internal() ApiError {
	return ApiError{
		Status:  http.StatusInternalServerError,
		Message: InternalServerError,
		Code:    INTERNAL_ERROR,
		Detail:  UnexpectedError,
		LogErr:  true,
	}
}

func ServiceUnavailable(message, detail string) ApiError {
	return ApiError{
		Status:  http.StatusServiceUnavailable,
		Message: message,
		Code:    SERVICE_UNAVAILABLE,
		Detail:  detail,
		LogErr:  true,
	}
}

func TooManyRequests(detail string) ApiError {
	return ApiError{
		Status:  http.StatusTooManyRequests,
		Message: "Too many requests",
		Code:    RATE_LIMITED,
		Detail:  detail,
	}
}

func UnauthorizedCredentials(message, detail string) ApiError {
	if message == "" {
		message = Unauthorized
	}
	return ApiError{
		Status:  http.StatusUnauthorized,
		Message: message,
		Code:    UNAUTHORIZED,
		Detail:  detail,
	}
}
