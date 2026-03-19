package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	// "runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultServiceName = "falzo-api"
	defaultEnvironment = "development"
)

func getRootDir() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	return wd
}

func SetupLogger() {
	rootDir := getRootDir()
	logDir := filepath.Join(rootDir, "logs")

	// Create the logs directory if it does not exist yet.
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	today := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logDir, today+".log")
	serviceName := getEnv("APP_NAME", defaultServiceName)
	environment := getEnv("APP_ENV", defaultEnvironment)

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		relative := strings.TrimPrefix(filepath.ToSlash(file), filepath.ToSlash(rootDir)+"/")
		if relative == file {
			relative = filepath.Base(file)
		}
		return fmt.Sprintf("%s:%d", relative, line)
	}

	// Writer 1: colored console output for local debugging.
	consoleWriter := newConsoleWriter(os.Stdout, true)

	// Writer 2: daily file log with 7-day retention.
	fileWriter := &lumberjack.Logger{
		Filename:  logFilePath,
		MaxAge:    7, // keep logs for 7 days
		LocalTime: true,
	}

	// Keep file logs readable too, but without ANSI colors.
	fileConsoleWriter := newConsoleWriter(fileWriter, false)

	// Write to both console and file using the same readable layout.
	multi := zerolog.MultiLevelWriter(consoleWriter, fileConsoleWriter)

	log.Logger = zerolog.New(multi).
		With().
		Str("service", serviceName).
		Str("env", environment).
		Timestamp().
		Caller(). // includes caller file, function, and line
		Logger()
}

func newConsoleWriter(out io.Writer, color bool) zerolog.ConsoleWriter {
	writer := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    !color,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			"service",
			"env",
			zerolog.CallerFieldName,
			zerolog.MessageFieldName,
		},
		FieldsOrder: []string{
			"addr",
			"method",
			"path",
			"status",
			"duration_ms",
			"bytes",
			"remote_ip",
			"request_id",
			"content_length",
			"signal",
			"phase",
			"timeout",
			"error",
		},
		FieldsExclude: []string{
			"service",
			"env",
		},
	}

	writer.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%-5s", i))
	}

	writer.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	writer.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	}

	writer.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%v", i)
	}

	return writer
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
