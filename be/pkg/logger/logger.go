package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"falzo-be/pkg/config"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultServiceName         = "falzo-api"
	defaultEnvironment         = "development"
	defaultLogDir              = "logs"
	defaultLogLevel            = "info"
	defaultLogMaxSizeMB        = 50
	defaultLogMaxBackups       = 5
	defaultLogMaxAgeDays       = 7
	defaultHTTPLogBodyMaxBytes = 4096
)

var httpRequestLog = For("http.request")

type Config struct {
	ServiceName   string
	Environment   string
	Level         string
	Pretty        bool
	LogDir        string
	SensitiveKeys []string
	MaxSizeMB     int
	MaxBackups    int
	MaxAgeDays    int
}

func SetupLogger(cfgs ...Config) {
	cfg := buildConfig(cfgs...)
	rootDir := mustGetWorkingDir()
	logDir := resolveLogDir(rootDir, cfg.LogDir)
	level := parseLevel(cfg.Level)

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	zerolog.SetGlobalLevel(level)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		relative := strings.TrimPrefix(filepath.ToSlash(file), filepath.ToSlash(rootDir)+"/")
		if relative == file {
			relative = filepath.Base(file)
		}
		return fmt.Sprintf("%s:%d", relative, line)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		LocalTime:  false,
		Compress:   false,
	}

	log.Logger = zerolog.New(newSensitiveDataWriter(buildWriter(cfg, fileWriter), cfg.SensitiveKeys)).
		Level(level).
		With().
		Str("service", cfg.ServiceName).
		Str("env", cfg.Environment).
		Timestamp().
		Caller().
		Logger()
}

func buildConfig(cfgs ...Config) Config {
	cfg := Config{
		ServiceName:   config.GetEnv("APP_NAME", defaultServiceName),
		Environment:   config.GetEnv("APP_ENV", defaultEnvironment),
		Level:         config.GetEnv("LOG_LEVEL", defaultLogLevel),
		Pretty:        config.GetBool("LOG_PRETTY", strings.EqualFold(config.GetEnv("APP_ENV", defaultEnvironment), "development")),
		LogDir:        config.GetEnv("LOG_DIR", defaultLogDir),
		SensitiveKeys: loadSensitiveKeyMarkers(config.GetEnv("LOG_SENSITIVE_KEYS", "")),
		MaxSizeMB:     config.GetInt("LOG_MAX_SIZE_MB", defaultLogMaxSizeMB),
		MaxBackups:    config.GetInt("LOG_MAX_BACKUPS", defaultLogMaxBackups),
		MaxAgeDays:    config.GetInt("LOG_MAX_AGE_DAYS", defaultLogMaxAgeDays),
	}

	if len(cfgs) == 0 {
		return cfg
	}

	override := cfgs[0]
	if override.ServiceName != "" {
		cfg.ServiceName = override.ServiceName
	}
	if override.Environment != "" {
		cfg.Environment = override.Environment
	}
	if override.Level != "" {
		cfg.Level = override.Level
	}
	if override.Pretty {
		cfg.Pretty = true
	}
	if override.LogDir != "" {
		cfg.LogDir = override.LogDir
	}
	if len(override.SensitiveKeys) > 0 {
		cfg.SensitiveKeys = override.SensitiveKeys
	}
	if override.MaxSizeMB > 0 {
		cfg.MaxSizeMB = override.MaxSizeMB
	}
	if override.MaxBackups > 0 {
		cfg.MaxBackups = override.MaxBackups
	}
	if override.MaxAgeDays > 0 {
		cfg.MaxAgeDays = override.MaxAgeDays
	}

	return cfg
}

func buildWriter(cfg Config, fileWriter io.Writer) io.Writer {
	if cfg.Pretty {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		return io.MultiWriter(consoleWriter, fileWriter)
	}

	return io.MultiWriter(os.Stdout, fileWriter)
}

func parseLevel(level string) zerolog.Level {
	parsed, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(level)))
	if err != nil {
		return zerolog.InfoLevel
	}

	return parsed
}

func resolveLogDir(rootDir, logDir string) string {
	if filepath.IsAbs(logDir) {
		return logDir
	}

	return filepath.Join(rootDir, logDir)
}

func mustGetWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}

	return wd
}

type statusRecorder struct {
	http.ResponseWriter
	status          int
	bytes           int
	captureBody     bool
	bodyCaptureSize int
	bodyTruncated   bool
	bodySample      bytes.Buffer
	errorLogged     bool
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.captureResponseBody(data)

	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func (r *statusRecorder) Flush() {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *statusRecorder) MarkErrorLogged() {
	r.errorLogged = true
}

func RequestLogger(next http.Handler) http.Handler {
	bodyCfg := loadBodyLogConfig()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestBody := capturedJSONBody{}
		if bodyCfg.Enabled {
			requestBody = captureRequestBody(r, bodyCfg.MaxBytes)
		}

		rec := &statusRecorder{
			ResponseWriter:  w,
			captureBody:     bodyCfg.Enabled,
			bodyCaptureSize: bodyCfg.MaxBytes,
		}

		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		if rec.errorLogged {
			return
		}
		if !shouldLogRequest(rec.status) {
			return
		}

		durationMs := float64(time.Since(start).Microseconds()) / 1000
		event := requestLogEvent(r.Context(), rec.status, r.URL.Path)
		if event == nil {
			return
		}
		event = event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rec.status).
			Float64("duration_ms", durationMs).
			Int("bytes", rec.bytes).
			Str("remote_ip", r.RemoteAddr)

		if r.ContentLength >= 0 {
			event = event.Int64("content_length", r.ContentLength)
		}

		if bodyCfg.Enabled {
			if requestBody.Payload != nil {
				event = event.Interface("request_body", requestBody.Payload)
			}
			if requestBody.Truncated {
				event = event.Bool("request_body_truncated", true)
			}
			if responseBody, ok := rec.responseBody(); ok {
				event = event.Interface("response_body", responseBody)
			}
			if rec.bodyTruncated {
				event = event.Bool("response_body_truncated", true)
			}
		}

		event.Msg("request completed")
	})
}

func shouldLogRequest(status int) bool {
	return status >= http.StatusBadRequest
}

func requestLogEvent(ctx context.Context, status int, path string) *zerolog.Event {
	if path == "/favicon.ico" {
		return httpRequestLog.event(ctx, zerolog.DebugLevel)
	}
	if status >= http.StatusInternalServerError {
		return httpRequestLog.event(ctx, zerolog.ErrorLevel)
	}
	if status >= http.StatusBadRequest {
		return httpRequestLog.event(ctx, zerolog.WarnLevel)
	}
	return httpRequestLog.event(ctx, zerolog.InfoLevel)
}

func sanitizeRequestID(requestID string) string {
	if _, suffix, found := strings.Cut(requestID, "/"); found {
		return suffix
	}

	return requestID
}

type bodyLogConfig struct {
	Enabled  bool
	MaxBytes int
}

func loadBodyLogConfig() bodyLogConfig {
	enabledByDefault := strings.EqualFold(config.GetEnv("APP_ENV", defaultEnvironment), "development")
	enabled := config.GetBool("LOG_HTTP_BODY_ENABLED", enabledByDefault)
	maxBytes := config.GetInt("LOG_HTTP_BODY_MAX_BYTES", defaultHTTPLogBodyMaxBytes)
	if maxBytes <= 0 {
		maxBytes = defaultHTTPLogBodyMaxBytes
	}

	return bodyLogConfig{
		Enabled:  enabled,
		MaxBytes: maxBytes,
	}
}

type capturedJSONBody struct {
	Payload   any
	Truncated bool
}

func captureRequestBody(r *http.Request, maxBytes int) capturedJSONBody {
	if r == nil || r.Body == nil || maxBytes <= 0 {
		return capturedJSONBody{}
	}

	originalBody := r.Body
	sample, err := io.ReadAll(io.LimitReader(originalBody, int64(maxBytes+1)))
	if err != nil {
		r.Body = originalBody
		return capturedJSONBody{}
	}
	r.Body = replayBody(sample, originalBody)

	if len(sample) == 0 {
		return capturedJSONBody{}
	}

	if len(sample) > maxBytes {
		return capturedJSONBody{Truncated: true}
	}

	payload, ok := decodeJSONPayload(sample)
	if !ok {
		return capturedJSONBody{}
	}

	return capturedJSONBody{Payload: payload}
}

func (r *statusRecorder) captureResponseBody(data []byte) {
	if r == nil || !r.captureBody || len(data) == 0 || r.bodyCaptureSize <= 0 || r.bodyTruncated {
		return
	}

	remaining := r.bodyCaptureSize - r.bodySample.Len()
	if remaining <= 0 {
		r.bodyTruncated = true
		return
	}

	if len(data) > remaining {
		_, _ = r.bodySample.Write(data[:remaining])
		r.bodyTruncated = true
		return
	}

	_, _ = r.bodySample.Write(data)
}

func (r *statusRecorder) responseBody() (any, bool) {
	if r == nil || r.bodySample.Len() == 0 {
		return nil, false
	}

	return decodeJSONPayload(r.bodySample.Bytes())
}

func decodeJSONPayload(data []byte) (any, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, false
	}

	var payload any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return nil, false
	}

	return payload, true
}

type replayReadCloser struct {
	io.Reader
	io.Closer
}

func replayBody(prefix []byte, body io.ReadCloser) io.ReadCloser {
	return replayReadCloser{
		Reader: io.MultiReader(bytes.NewReader(prefix), body),
		Closer: body,
	}
}
