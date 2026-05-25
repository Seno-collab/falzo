package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	origins := normalizeValues(cfg.AllowedOrigins)
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	methods := normalizeMethods(cfg.AllowedMethods)
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	headers := normalizeValues(cfg.AllowedHeaders)
	if len(headers) == 0 {
		headers = []string{"Accept", "Authorization", "Content-Type", "Origin", "X-Requested-With"}
	}

	allowAnyOrigin := containsFold(origins, "*") && !cfg.AllowCredentials
	allowMethodsValue := strings.Join(methods, ", ")
	allowHeadersValue := strings.Join(headers, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !allowAnyOrigin && !containsFold(origins, origin) {
				if isPreflightRequest(r) {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			addVary(w.Header(), "Origin")
			setAllowOrigin(w.Header(), allowAnyOrigin, origin)

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if !isPreflightRequest(r) {
				next.ServeHTTP(w, r)
				return
			}

			addVary(w.Header(), "Access-Control-Request-Method")
			addVary(w.Header(), "Access-Control-Request-Headers")

			requestMethod := strings.ToUpper(strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")))
			if requestMethod != "" && !containsFold(methods, requestMethod) {
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}

			requestHeaders := normalizeValues(strings.Split(r.Header.Get("Access-Control-Request-Headers"), ","))
			if len(requestHeaders) > 0 && !areHeadersAllowed(requestHeaders, headers) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			w.Header().Set("Access-Control-Allow-Methods", allowMethodsValue)
			w.Header().Set("Access-Control-Allow-Headers", allowHeadersValue)
			if cfg.MaxAgeSeconds > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAgeSeconds))
			}

			w.WriteHeader(http.StatusNoContent)
		})
	}
}

func isPreflightRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions && strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")) != ""
}

func setAllowOrigin(header http.Header, allowAnyOrigin bool, origin string) {
	if allowAnyOrigin {
		header.Set("Access-Control-Allow-Origin", "*")
		return
	}

	header.Set("Access-Control-Allow-Origin", origin)
}

func areHeadersAllowed(requested []string, allowed []string) bool {
	if containsFold(allowed, "*") {
		return true
	}

	for _, header := range requested {
		if !containsFold(allowed, header) {
			return false
		}
	}

	return true
}

func normalizeMethods(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.ToUpper(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func normalizeValues(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func addVary(header http.Header, value string) {
	existing := header.Values("Vary")
	for _, current := range existing {
		for _, part := range strings.Split(current, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}

	header.Add("Vary", value)
}
