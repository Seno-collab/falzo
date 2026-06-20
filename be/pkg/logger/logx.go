package logger

import (
	"context"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	category string
	fields   []Field
}

type Field func(*zerolog.Event)

func For(category string) Logger {
	return Logger{category: category}
}

func (l Logger) With(fields ...Field) Logger {
	next := Logger{
		category: l.category,
		fields:   make([]Field, 0, len(l.fields)+len(fields)),
	}
	next.fields = append(next.fields, l.fields...)
	next.fields = append(next.fields, fields...)
	return next
}

func (l Logger) Debug(ctx context.Context, message string, fields ...Field) {
	l.write(ctx, zerolog.DebugLevel, nil, message, fields...)
}

func (l Logger) Info(ctx context.Context, message string, fields ...Field) {
	l.write(ctx, zerolog.InfoLevel, nil, message, fields...)
}

func (l Logger) Warn(ctx context.Context, err error, message string, fields ...Field) {
	l.write(ctx, zerolog.WarnLevel, err, message, fields...)
}

func (l Logger) Error(ctx context.Context, err error, message string, fields ...Field) {
	l.write(ctx, zerolog.ErrorLevel, err, message, fields...)
}

func (l Logger) Fatal(ctx context.Context, err error, message string, fields ...Field) {
	l.write(ctx, zerolog.FatalLevel, err, message, fields...)
}

func (l Logger) write(ctx context.Context, level zerolog.Level, err error, message string, fields ...Field) {
	event := l.event(ctx, level)
	if event == nil {
		return
	}

	if err != nil {
		event = event.Err(err)
	}
	applyFields(event, fields...)
	event.Msg(message)
}

func (l Logger) event(ctx context.Context, level zerolog.Level) *zerolog.Event {
	var event *zerolog.Event
	switch level {
	case zerolog.DebugLevel:
		event = log.Debug()
	case zerolog.InfoLevel:
		event = log.Info()
	case zerolog.WarnLevel:
		event = log.Warn()
	case zerolog.ErrorLevel:
		event = log.Error()
	case zerolog.FatalLevel:
		event = log.Fatal()
	default:
		event = log.WithLevel(level)
	}
	if event == nil {
		return nil
	}

	if l.category != "" {
		event = event.Str("category", l.category)
	}
	if ctx != nil {
		if requestID := chimiddleware.GetReqID(ctx); requestID != "" {
			event = event.Str("request_id", sanitizeRequestID(requestID))
		}
	}
	applyFields(event, l.fields...)
	return event
}

func applyFields(event *zerolog.Event, fields ...Field) {
	if event == nil {
		return
	}
	for _, field := range fields {
		if field != nil {
			field(event)
		}
	}
}

func Str(key, value string) Field {
	return func(event *zerolog.Event) {
		event.Str(key, value)
	}
}

func Strs(key string, value []string) Field {
	return func(event *zerolog.Event) {
		event.Strs(key, value)
	}
}

func Int(key string, value int) Field {
	return func(event *zerolog.Event) {
		event.Int(key, value)
	}
}

func Int64(key string, value int64) Field {
	return func(event *zerolog.Event) {
		event.Int64(key, value)
	}
}

func Uint64(key string, value uint64) Field {
	return func(event *zerolog.Event) {
		event.Uint64(key, value)
	}
}

func Float64(key string, value float64) Field {
	return func(event *zerolog.Event) {
		event.Float64(key, value)
	}
}

func Bool(key string, value bool) Field {
	return func(event *zerolog.Event) {
		event.Bool(key, value)
	}
}

func Dur(key string, value time.Duration) Field {
	return func(event *zerolog.Event) {
		event.Dur(key, value)
	}
}

func Interface(key string, value any) Field {
	return func(event *zerolog.Event) {
		event.Interface(key, value)
	}
}

func AnErr(key string, err error) Field {
	return func(event *zerolog.Event) {
		event.AnErr(key, err)
	}
}
