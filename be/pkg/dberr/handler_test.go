package dberr

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

var _ net.Error = timeoutErr{}

func TestClassifyUniqueViolation(t *testing.T) {
	kind, code := Classify(&pgconn.PgError{Code: "23505"})
	if kind != KindUnique {
		t.Fatalf("expected kind %q, got %q", KindUnique, kind)
	}
	if code != "23505" {
		t.Fatalf("expected code 23505, got %q", code)
	}
}

func TestClassifyForeignKeyViolation(t *testing.T) {
	kind, code := Classify(&pgconn.PgError{Code: "23503"})
	if kind != KindForeignKey {
		t.Fatalf("expected kind %q, got %q", KindForeignKey, kind)
	}
	if code != "23503" {
		t.Fatalf("expected code 23503, got %q", code)
	}
}

func TestClassifyDependencyErrors(t *testing.T) {
	tests := []error{
		context.DeadlineExceeded,
		sql.ErrConnDone,
		timeoutErr{},
		errors.New("dial tcp 10.0.0.1:5432: connect: connection refused"),
	}

	for _, err := range tests {
		kind, _ := Classify(err)
		if kind != KindDependency {
			t.Fatalf("expected dependency kind for %v, got %q", err, kind)
		}
	}
}

func TestHandleUsesMapper(t *testing.T) {
	expected := errors.New("mapped")
	got := Handle(errors.New("boom"), "auth", "accounts.find", "req-1", func(kind Kind, err error) error {
		if kind != KindInternal {
			t.Fatalf("expected internal kind, got %q", kind)
		}
		return expected
	})

	if !errors.Is(got, expected) {
		t.Fatalf("expected mapped error, got %v", got)
	}
}

func TestHandleDoesNotEmitLog(t *testing.T) {
	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})
	log.Logger = zerolog.New(&output)

	_ = Handle(errors.New("boom"), "auth", "accounts.find", "req-1", nil)

	if got := strings.TrimSpace(output.String()); got != "" {
		t.Fatalf("expected database handler not to emit logs, got %q", got)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if !IsUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("expected unique violation to be true")
	}
	if IsUniqueViolation(errors.New("not unique")) {
		t.Fatal("expected unique violation to be false")
	}
}

func TestMapDependencyOrInternal(t *testing.T) {
	dependencyErr := errors.New("dependency unavailable")
	internalErr := errors.New("internal error")

	got := MapDependencyOrInternal(
		context.DeadlineExceeded,
		"auth",
		"accounts.find",
		"req-1",
		dependencyErr,
		internalErr,
	)
	if !errors.Is(got, dependencyErr) {
		t.Fatalf("expected dependency error, got %v", got)
	}
	if !errors.Is(got, context.DeadlineExceeded) {
		t.Fatalf("expected dependency cause to be preserved, got %v", got)
	}

	got = MapDependencyOrInternal(
		errors.New("boom"),
		"auth",
		"accounts.find",
		"req-2",
		dependencyErr,
		internalErr,
	)
	if !errors.Is(got, internalErr) {
		t.Fatalf("expected internal error, got %v", got)
	}
}
