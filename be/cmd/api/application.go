package main

import (
	roomapp "be/internal/application/room"
	"be/internal/realtime"
	"log/slog"
	"net/http"
)

// application contains the long-lived components produced by Wire. Startup and
// shutdown stay explicit in main.go because they depend on process lifecycle.
type application struct {
	logger         *slog.Logger
	server         *http.Server
	realtimeHub    *realtime.Hub
	phaseScheduler *roomapp.PhaseScheduler
}
