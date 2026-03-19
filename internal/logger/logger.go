package logger

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Writer 1: stdout JSON for containers and centralized logging.
	stdoutWriter := os.Stdout

	// Writer 2: daily file log with 7-day retention.
	fileWriter := &lumberjack.Logger{
		Filename:  logFilePath,
		MaxAge:    7, // keep logs for 7 days
		LocalTime: true,
	}

	// Write JSON to both stdout and file.
	multi := zerolog.MultiLevelWriter(stdoutWriter, fileWriter)

	log.Logger = zerolog.New(multi).
		With().
		Str("service", serviceName).
		Str("env", environment).
		Timestamp().
		Caller(). // includes caller file, function, and line
		Logger()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
