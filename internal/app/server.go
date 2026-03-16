package app

import (
	"context"
	"falzo/pkg/lib"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Run() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
	sm := lib.NewShutdownManager(60 * time.Second)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	sm.Register("http-stop-acceting", 5*time.Second, func(ctx context.Context) error {
		// Stop accepting new requests immediately
		return srv.Shutdown(ctx)
	})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("server failed", "error", err)
			os.Exit(1)
		}
	}()
	fmt.Println("server started", "addr", srv.Addr)
	sig := <-quit
	fmt.Println("shutdown initiated", "signal", sig)

	if err := sm.Shutdown(); err != nil {
		fmt.Println("shutdown completed with errors", "error", err)
		os.Exit(1)
	}

	fmt.Println("shutdown complete")
}
