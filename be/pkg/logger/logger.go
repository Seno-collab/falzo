package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultServiceName   = "falzo-api"
	defaultEnvironment   = "development"
	defaultLogDir        = "logs"
	defaultLogLevel      = "info"
	defaultLogMaxSizeMB  = 50
	defaultLogMaxBackups = 5
	defaultLogMaxAgeDays = 7
)

type Config struct {
	ServiceName string
	Environment string
	Level       string
	Pretty      bool
	LogDir      string
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
}

func SetupLogger(cfgs ...Config) {
	cfg := buildConfig(cfgs...)
	rootDir := mustGetWorkingDir()
	logDir := resolveLogDir(rootDir, cfg.LogDir)
	level := parseLevel(cfg.Level)

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
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
		LocalTime:  true,
		Compress:   false,
	}

	log.Logger = zerolog.New(buildWriter(cfg, fileWriter)).
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
		ServiceName: getEnv("APP_NAME", defaultServiceName),
		Environment: getEnv("APP_ENV", defaultEnvironment),
		Level:       getEnv("LOG_LEVEL", defaultLogLevel),
		Pretty:      getBool("LOG_PRETTY", strings.EqualFold(getEnv("APP_ENV", defaultEnvironment), "development")),
		LogDir:      getEnv("LOG_DIR", defaultLogDir),
		MaxSizeMB:   getInt("LOG_MAX_SIZE_MB", defaultLogMaxSizeMB),
		MaxBackups:  getInt("LOG_MAX_BACKUPS", defaultLogMaxBackups),
		MaxAgeDays:  getInt("LOG_MAX_AGE_DAYS", defaultLogMaxAgeDays),
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
			TimeFormat: "2006-01-02 15:04:05",
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

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		event := log.Info()
		if r.URL.Path == "/favicon.ico" {
			event = log.Debug()
		}

		durationMs := float64(time.Since(start).Microseconds()) / 1000
		event = event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rec.status).
			Float64("duration_ms", durationMs).
			Int("bytes", rec.bytes).
			Str("remote_ip", r.RemoteAddr)

		if requestID := chimiddleware.GetReqID(r.Context()); requestID != "" {
			event = event.Str("request_id", sanitizeRequestID(requestID))
		}

		if r.ContentLength >= 0 {
			event = event.Int64("content_length", r.ContentLength)
		}

		event.Msg("request completed")
	})
}

func sanitizeRequestID(requestID string) string {
	if _, suffix, found := strings.Cut(requestID, "/"); found {
		return suffix
	}

	return requestID
}
