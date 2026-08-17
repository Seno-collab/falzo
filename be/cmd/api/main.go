package main

import (
	"be/internal/alerting"
	"be/internal/api"
	"be/internal/api/handler"
	apimiddleware "be/internal/api/middleware"
	authapp "be/internal/application/auth"
	chatapp "be/internal/application/chat"
	roomapp "be/internal/application/room"
	socialapp "be/internal/application/social"
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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
	if err := run(); err != nil {
		slog.Error("api stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() (runErr error) {
	cfg, err := config.Load(".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	appLogger, logCloser, err := loggerx.New(cfg.Server.Env, "logs")
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer logCloser.Close()
	slog.SetDefault(appLogger)
	slog.Info("api starting", slog.String("address", ":"+cfg.Server.Port))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	metrics := observability.NewMetrics(nil)
	if cfg.NATS.Enabled {
		alertPublisher := natsinfra.NewAlertPublisher(
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
		slog.SetDefault(appLogger)
		defer func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := alertPublisher.Close(closeCtx); err != nil {
				appLogger.Warn("close NATS alert publisher", slog.Any("error", err))
			}
		}()
		defer func() {
			if runErr != nil {
				appLogger.Error("api stopped", slog.Any("error", runErr))
			}
		}()
	}
	db, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	redisClient, err := redisinfra.NewRedisClient(ctx, cfg.Redis, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()

	c := clock.NewSystemClock()
	users := postgres.NewUserRepository(db)
	hasher := security.NewBcryptPasswordHasher(12)
	tokens, err := security.NewJWTTokenManager(cfg.JWT.Secret, time.Duration(cfg.JWT.ExpiredMinutes)*time.Minute, time.Duration(cfg.JWT.RefreshExpiredHours)*time.Hour, time.Duration(cfg.JWT.ResetExpiredMinutes)*time.Minute)
	if err != nil {
		return fmt.Errorf("create token manager: %w", err)
	}
	sessions := redisinfra.NewTokenSessionStore(redisClient, cfg.Redis.KeyPrefix)
	login := authapp.NewLoginUseCase(hasher, users, tokens, sessions, cfg.Auth.MaxLoginAttempts, time.Duration(cfg.Auth.LockMinutes)*time.Minute, c)
	refresh := authapp.NewRefreshTokenUseCase(users, tokens, sessions, c)
	forgot := authapp.NewForgotPasswordUseCase(users, tokens, sessions, c)
	reset := authapp.NewResetPasswordUseCase(users, hasher, tokens, sessions, c)
	logout := authapp.NewLogoutUseCase(tokens, sessions, c)
	googleVerifier := security.NewGoogleIDTokenVerifier(cfg.Google.ClientID)
	googleLogin := authapp.NewGoogleLoginUseCase(googleVerifier, users, tokens, sessions, c)
	authHandler := handler.NewAuthHandler(login, refresh, forgot, reset, logout, cfg.Server.Env != "production", appLogger, googleLogin)
	rooms := postgres.NewRoomRepository(db)
	inviteCodes := security.NewInviteCodeGenerator(6)
	createRoom := roomapp.NewCreateRoomUseCase(
		rooms,
		inviteCodes,
		c,
		cfg.Game.MaxPlayersPerRoom,
		time.Duration(cfg.Game.RoomTimeoutSeconds)*time.Second,
	)
	listRooms := roomapp.NewListRoomsUseCase(rooms)
	getRoom := roomapp.NewGetRoomUseCase(rooms)
	joinRoom := roomapp.NewJoinRoomUseCase(rooms)
	kickMember := roomapp.NewKickMemberUseCase(rooms)
	dealRound := roomapp.NewDealRoundUseCase(rooms, c)
	getCurrentCard := roomapp.NewGetCurrentCardUseCase(rooms)
	updateDiscussion := roomapp.NewUpdateDiscussionUseCase(rooms)
	getRoundState := roomapp.NewGetRoundStateUseCase(rooms, c)
	castVote := roomapp.NewCastVoteUseCase(rooms, c)
	playerReady := roomapp.NewPlayerReadyUseCase(rooms, c)
	finishTurn := roomapp.NewFinishTurnUseCase(rooms, c)
	mrWhiteGuess := roomapp.NewMrWhiteGuessUseCase(rooms, c)
	chatRepository := postgres.NewChatRepository(db)
	chatService := chatapp.NewService(chatRepository)
	realtimeBackplane := redisinfra.NewRealtimeBackplane(redisClient, cfg.Redis.KeyPrefix, appLogger, metrics)
	realtimeHub := realtime.NewHub(
		c,
		realtime.WithBackplane(realtimeBackplane),
		realtime.WithLogger(appLogger),
		realtime.WithMetrics(metrics),
		realtime.WithChatStore(chatService),
	)
	if err := realtimeHub.Start(ctx); err != nil {
		return fmt.Errorf("start realtime hub: %w", err)
	}
	defer realtimeHub.Close()
	roomHandler := handler.NewRoomHandler(
		createRoom,
		listRooms,
		getRoom,
		joinRoom,
		kickMember,
		dealRound,
		getCurrentCard,
		updateDiscussion,
		getRoundState,
		castVote,
		playerReady,
		finishTurn,
		mrWhiteGuess,
		realtimeHub,
		appLogger,
	)
	realtimeHandler := handler.NewRealtimeHandler(
		getRoom,
		getRoundState,
		realtimeHub,
		cfg.Server.WebSocketOriginPatterns,
		appLogger,
		metrics,
	)
	chatHandler := handler.NewChatHandler(chatService, getRoom, appLogger)
	phaseScheduler := roomapp.NewPhaseScheduler(rooms, c, metrics, appLogger, func(ctx context.Context, transition domainroom.PhaseTransition) {
		realtimeHub.Publish(transition.RoomID, realtime.EventStateUpdated, realtime.StateUpdated{
			Round:             transition.RoundNumber,
			Cycle:             transition.CycleNumber,
			Phase:             transition.To,
			CurrentTurnUserID: transition.CurrentTurnPlayerID,
			PhaseDeadlineAt:   transition.PhaseDeadlineAt,
		})
		if !transition.MembersChanged {
			return
		}
		room, err := getRoom.Execute(ctx, transition.RoomID)
		if err != nil {
			appLogger.WarnContext(ctx, "refresh scheduled room members", slog.Any("error", err))
			return
		}
		realtimeHub.UpdateMembers(room.ID, realtimeMembers(room))
		realtimeHub.Publish(room.ID, realtime.EventRoomUpdated, realtime.RoomUpdated{
			Version: room.Version,
			Reason:  "phase_transition",
		})
	})
	phaseScheduler.Start(ctx)
	socialRepository := postgres.NewSocialRepository(db)
	socialService := socialapp.NewService(socialRepository, c)
	socialHandler := handler.NewSocialHandler(socialService, appLogger, realtimeHub)
	healthHandler := handler.NewHealthHandler(
		func(ctx context.Context) error { return db.Ping(ctx) },
		func(ctx context.Context) error { return redisClient.Ping(ctx).Err() },
	)
	authenticator := apimiddleware.NewAuthenticator(tokens)

	server := &http.Server{Addr: ":" + cfg.Server.Port, Handler: api.NewRouter(healthHandler, authHandler, roomHandler, realtimeHandler, chatHandler, socialHandler, authenticator, promhttp.Handler(), appLogger), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	slog.Info("api listening", slog.String("address", server.Addr))
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("api shutting down", slog.String("reason", ctx.Err().Error()))
		realtimeHub.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
	return nil
}

func realtimeMembers(room *domainroom.Room) []realtime.Member {
	members := make([]realtime.Member, 0, len(room.Members))
	for _, member := range room.Members {
		members = append(members, realtime.Member{
			ID: member.UserID, Name: member.UserName, SeatNumber: member.SeatNumber,
			Host: member.IsHost, Eliminated: member.Eliminated,
		})
	}
	return members
}
