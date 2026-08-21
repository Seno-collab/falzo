package main

import (
	"be/internal/alerting"
	"be/internal/api/handler"
	authapp "be/internal/application/auth"
	chatapp "be/internal/application/chat"
	authports "be/internal/application/ports/auth"
	chatports "be/internal/application/ports/chat"
	roomports "be/internal/application/ports/room"
	socialports "be/internal/application/ports/social"
	roomapp "be/internal/application/room"
	"be/internal/config"
	domainroom "be/internal/domain/room"
	natsinfra "be/internal/insfrastructure/nats"
	"be/internal/insfrastructure/persistence/postgres"
	redisinfra "be/internal/insfrastructure/redis"
	"be/internal/insfrastructure/security"
	"be/internal/observability"
	"be/internal/realtime"
	"be/internal/shared/clock"
	loggerx "be/internal/shared/logger"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goredis "github.com/redis/go-redis/v9"
)

func provideMetrics() *observability.Metrics {
	return observability.NewMetrics(nil)
}

func provideLogger(cfg *config.Config, metrics *observability.Metrics) (*slog.Logger, func(), error) {
	appLogger, logCloser, err := loggerx.New(cfg.Server.Env, "logs")
	if err != nil {
		return nil, nil, err
	}

	var alertPublisher *natsinfra.AlertPublisher
	if cfg.NATS.Enabled {
		alertPublisher = natsinfra.NewAlertPublisher(
			cfg.NATS.URL,
			cfg.NATS.Stream,
			cfg.NATS.Subject,
			metrics,
		)
		appLogger = slog.New(alerting.NewSlogHandler(
			appLogger.Handler(),
			alertPublisher,
			"falzo-api",
			cfg.Server.Env,
		))
	}

	cleanup := func() {
		if alertPublisher != nil {
			closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := alertPublisher.Close(closeCtx); err != nil {
				appLogger.Warn("close NATS alert publisher", slog.Any("error", err))
			}
			cancel()
		}
		_ = logCloser.Close()
	}
	return appLogger, cleanup, nil
}

func provideDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, func(), error) {
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, nil, err
	}
	return pool, pool.Close, nil
}

func provideRedis(ctx context.Context, cfg *config.Config) (*goredis.Client, func(), error) {
	client, err := redisinfra.NewRedisClient(ctx, cfg.Redis, cfg.Redis.DB)
	if err != nil {
		return nil, nil, err
	}
	return client, func() { _ = client.Close() }, nil
}

func provideClock() clock.Clock {
	return clock.NewSystemClock()
}

func provideUserRepository(db *pgxpool.Pool) authports.UserRepository {
	return postgres.NewUserRepository(db)
}

func provideRoomRepository(db *pgxpool.Pool) roomports.Repository {
	return postgres.NewRoomRepository(db)
}

func provideChatRepository(db *pgxpool.Pool) chatports.Repository {
	return postgres.NewChatRepository(db)
}

func provideSocialRepository(db *pgxpool.Pool) socialports.Repository {
	return postgres.NewSocialRepository(db)
}

func providePasswordHasher() authports.PasswordHasher {
	return security.NewBcryptPasswordHasher(12)
}

func provideTokenManager(cfg *config.Config) (authports.TokenManager, error) {
	return security.NewJWTTokenManager(
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.ExpiredMinutes)*time.Minute,
		time.Duration(cfg.JWT.RefreshExpiredHours)*time.Hour,
		time.Duration(cfg.JWT.ResetExpiredMinutes)*time.Minute,
	)
}

func provideTokenSessionStore(client *goredis.Client, cfg *config.Config) authports.TokenSessionStore {
	return redisinfra.NewTokenSessionStore(client, cfg.Redis.KeyPrefix)
}

func provideGoogleIdentityVerifier(cfg *config.Config) authports.GoogleIdentityVerifier {
	return security.NewGoogleIDTokenVerifier(cfg.Google.ClientID)
}

func provideInviteCodeGenerator() roomports.InviteCodeGenerator {
	return security.NewInviteCodeGenerator(6)
}

func provideLoginUseCase(
	hasher authports.PasswordHasher,
	users authports.UserRepository,
	tokens authports.TokenManager,
	sessions authports.TokenSessionStore,
	cfg *config.Config,
	c clock.Clock,
) *authapp.LoginUseCase {
	return authapp.NewLoginUseCase(
		hasher,
		users,
		tokens,
		sessions,
		cfg.Auth.MaxLoginAttempts,
		time.Duration(cfg.Auth.LockMinutes)*time.Minute,
		c,
	)
}

func provideCreateRoomUseCase(
	rooms roomports.Repository,
	inviteCodes roomports.InviteCodeGenerator,
	c clock.Clock,
	cfg *config.Config,
) *roomapp.CreateRoomUseCase {
	return roomapp.NewCreateRoomUseCase(
		rooms,
		inviteCodes,
		c,
		cfg.Game.MaxPlayersPerRoom,
		time.Duration(cfg.Game.RoomTimeoutSeconds)*time.Second,
	)
}

func provideRealtimeBackplane(
	client *goredis.Client,
	cfg *config.Config,
	logger *slog.Logger,
	metrics *observability.Metrics,
) realtime.Backplane {
	return redisinfra.NewRealtimeBackplane(client, cfg.Redis.KeyPrefix, logger, metrics)
}

func provideRealtimeHub(
	c clock.Clock,
	backplane realtime.Backplane,
	logger *slog.Logger,
	metrics *observability.Metrics,
	chatService *chatapp.Service,
) *realtime.Hub {
	return realtime.NewHub(
		c,
		realtime.WithBackplane(backplane),
		realtime.WithLogger(logger),
		realtime.WithMetrics(metrics),
		realtime.WithChatStore(chatService),
	)
}

func provideAuthHandler(
	register *authapp.RegisterUseCase,
	login *authapp.LoginUseCase,
	refresh *authapp.RefreshTokenUseCase,
	forgot *authapp.ForgotPasswordUseCase,
	reset *authapp.ResetPasswordUseCase,
	logout *authapp.LogoutUseCase,
	logger *slog.Logger,
	googleLogin *authapp.GoogleLoginUseCase,
	cfg *config.Config,
) *handler.AuthHandler {
	return handler.NewAuthHandler(
		register,
		login,
		refresh,
		forgot,
		reset,
		logout,
		cfg.Server.Env != "production",
		logger,
		googleLogin,
	)
}

func provideRealtimeHandler(
	getRoom *roomapp.GetRoomUseCase,
	getRoundState *roomapp.GetRoundStateUseCase,
	hub *realtime.Hub,
	logger *slog.Logger,
	metrics *observability.Metrics,
	cfg *config.Config,
) *handler.RealtimeHandler {
	return handler.NewRealtimeHandler(
		getRoom,
		getRoundState,
		hub,
		cfg.Server.WebSocketOriginPatterns,
		logger,
		metrics,
	)
}

func providePhaseTransitionHandler(
	hub *realtime.Hub,
	getRoom *roomapp.GetRoomUseCase,
	logger *slog.Logger,
) roomapp.PhaseTransitionHandler {
	return func(ctx context.Context, transition domainroom.PhaseTransition) {
		hub.Publish(transition.RoomID, realtime.EventStateUpdated, realtime.StateUpdated{
			Round:             transition.RoundNumber,
			Cycle:             transition.CycleNumber,
			Phase:             transition.To,
			CurrentTurnUserID: transition.CurrentTurnPlayerID,
			TurnNumber:        transition.TurnNumber,
			TotalTurns:        transition.TotalTurns,
			TurnEndsAt:        transition.TurnEndsAt,
			PhaseDeadlineAt:   transition.PhaseDeadlineAt,
		})
		if !transition.MembersChanged {
			return
		}
		room, err := getRoom.Execute(ctx, transition.RoomID)
		if err != nil {
			logger.WarnContext(ctx, "refresh scheduled room members", slog.Any("error", err))
			return
		}
		hub.UpdateMembers(room.ID, realtimeMembers(room))
		hub.Publish(room.ID, realtime.EventRoomUpdated, realtime.RoomUpdated{
			Version: room.Version,
			Reason:  "phase_transition",
		})
	}
}

func provideHealthHandler(db *pgxpool.Pool, redisClient *goredis.Client) *handler.HealthHandler {
	return handler.NewHealthHandler(
		func(ctx context.Context) error { return db.Ping(ctx) },
		func(ctx context.Context) error { return redisClient.Ping(ctx).Err() },
	)
}

func provideMetricsHandler() http.Handler {
	return promhttp.Handler()
}

func provideHTTPServer(cfg *config.Config, router *chi.Mux) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func realtimeMembers(room *domainroom.Room) []realtime.Member {
	members := make([]realtime.Member, 0, len(room.Members))
	for _, member := range room.Members {
		members = append(members, realtime.Member{
			ID:         member.UserID,
			Name:       member.UserName,
			SeatNumber: member.SeatNumber,
			Host:       member.IsHost,
			Eliminated: member.Eliminated,
		})
	}
	return members
}
