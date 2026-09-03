package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates the application logger.
// Development logs are human-readable; deployed environments emit JSON so
// Docker, log collectors, and the daily log files can parse fields reliably.
func New(environment, logDir string) (*slog.Logger, io.Closer, error) {
	level := slog.LevelInfo
	if environment == "development" {
		level = slog.LevelDebug
	}

	fileWriter, err := newDailyFileWriter(logDir, "api")
	if err != nil {
		return nil, nil, err
	}

	options := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}
	var handler slog.Handler
	output := io.MultiWriter(os.Stdout, fileWriter)
	if environment == "development" {
		handler = slog.NewTextHandler(output, options)
	} else {
		handler = slog.NewJSONHandler(output, options)
	}

	return slog.New(handler).With(
		slog.String("service", "falzo-api"),
		slog.String("environment", environment),
	), fileWriter, nil
}
