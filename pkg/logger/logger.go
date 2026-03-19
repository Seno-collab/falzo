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

	stdoutWriter := os.Stdout
	fileWriter := &lumberjack.Logger{
		Filename:  logFilePath,
		MaxAge:    7,
		LocalTime: true,
	}

	multi := zerolog.MultiLevelWriter(stdoutWriter, fileWriter)

	log.Logger = zerolog.New(multi).
		With().
		Str("service", serviceName).
		Str("env", environment).
		Timestamp().
		Caller().
		Logger()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
