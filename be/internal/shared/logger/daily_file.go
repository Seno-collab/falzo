package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type dailyFileWriter struct {
	mu     sync.Mutex
	dir    string
	prefix string
	date   string
	file   *os.File
	closed bool
}

func newDailyFileWriter(dir, prefix string) (*dailyFileWriter, error) {
	if dir == "" {
		dir = "logs"
	}
	writer := &dailyFileWriter{dir: dir, prefix: prefix}
	if err := writer.rotateLocked(time.Now()); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *dailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, fmt.Errorf("log writer is closed")
	}
	if currentDate := dateOf(time.Now()); currentDate != w.date {
		if err := w.rotateLocked(time.Now()); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *dailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true
	return w.file.Close()
}

func (w *dailyFileWriter) rotateLocked(now time.Time) error {
	if err := os.MkdirAll(w.dir, 0o750); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	date := dateOf(now)
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close previous log file: %w", err)
		}
	}

	path := filepath.Join(w.dir, fmt.Sprintf("%s-%s.log", w.prefix, date))
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	w.file = file
	w.date = date
	return nil
}

func dateOf(t time.Time) string {
	return t.Format("2006-01-02")
}
