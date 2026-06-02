package app

import (
	"context"
	"falzo-be/internal/auth"
	authInfra "falzo-be/internal/auth/infra"
	"falzo-be/internal/category"
	categoryInfra "falzo-be/internal/category/infra"
	"falzo-be/internal/location"
	locationInfra "falzo-be/internal/location/infra"
	"falzo-be/internal/notification"
	notificationInfra "falzo-be/internal/notification/infra"
	"falzo-be/internal/post"
	postInfra "falzo-be/internal/post/infra"
	"falzo-be/internal/social"
	socialInfra "falzo-be/internal/social/infra"
	"falzo-be/internal/upload"
	uploadInfra "falzo-be/internal/upload/infra"
	"falzo-be/pkg/cache"
	"falzo-be/pkg/config"
	"falzo-be/pkg/database"
	httpMiddleware "falzo-be/pkg/http/middleware"
	"falzo-be/pkg/logger"
	"falzo-be/pkg/shutdown"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type authAvatarEventPublisher struct {
	posts    post.PostEventPublisher
	profiles publicProfileCacheInvalidator
}

type publicProfileCacheInvalidator interface {
	InvalidatePublicProfile(ctx context.Context, userID uint64)
}

func (p authAvatarEventPublisher) PublishAvatarUpdated(ctx context.Context, event auth.AvatarUpdatedEvent) error {
	if p.profiles != nil {
		p.profiles.InvalidatePublicProfile(ctx, event.UserID)
	}
	if p.posts == nil {
		return nil
	}

	return p.posts.PublishUserAvatarUpdated(ctx, post.UserAvatarUpdatedEvent{
		UserID:         event.UserID,
		AvatarURL:      event.AvatarURL,
		AvatarURLAlias: event.AvatarURLAlias,
	})
}

func Run() {
	// Keep application-wide local time aligned to UTC (UTC+0).
	time.Local = time.UTC

	config.BootstrapEnv()
	logger.SetupLogger()
	cfg := config.Load()
	if err := config.Validate(cfg); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	db, err := database.New(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect postgres")
	}

	accounts := authInfra.NewAccountRepository(db)
	var sessions auth.SessionRepository = authInfra.NewSessionRepository(db)
	redisClient, err := cache.New(cfg.Redis)
	if err != nil {
		log.Warn().Err(err).Msg("redis unavailable, continuing without session cache")
	} else {
		sessions = authInfra.NewCachedSessionRepository(sessions, redisClient, cfg.Auth.TokenTTL)
	}
	passwords := authInfra.NewPasswordHasher()
	jwtManager := authInfra.NewJWTManager(cfg.Auth)
	authService := auth.NewService(accounts, sessions, passwords, jwtManager, jwtManager, cfg.Auth.RefreshTokenTTL)
	sessionCleanupCtx, stopSessionCleanup := context.WithCancel(context.Background())
	go auth.RunSessionCleanup(sessionCleanupCtx, sessions)
	authRateLimit := httpMiddleware.NewIPRateLimiter(cfg.Auth.RateLimitPerMin, time.Minute)
	readRateLimit := httpMiddleware.NewIPRateLimiter(cfg.HTTP.ReadRateLimitPerMin, time.Minute)
	authProtector := auth.WithProtectorConfig(cfg.Auth.RateLimitPerMin, cfg.Auth.DependencyFailureThreshold, cfg.Auth.DependencyCoolDown)
	authHandler := auth.NewHandler(authService, authProtector, auth.WithPublicMiddlewares(authRateLimit))
	locationRepository := locationInfra.NewPostgresRepository(db)
	locationService := location.NewService(locationRepository)
	locationHandler := location.NewHandler(locationService, location.WithReadMiddlewares(readRateLimit))
	notificationRepository := notificationInfra.NewPostgresRepository(db)
	notificationHub := notification.NewHub(notificationRepository)
	notificationHandler := notification.NewHandler(notificationHub, authService)
	grpcServer := grpc.NewServer()
	notification.RegisterNotificationServiceServer(grpcServer, notification.NewGRPCServer(notificationHub, authService))
	var socialRepository social.Repository = socialInfra.NewPostgresRepository(db)
	if redisClient != nil {
		socialRepository = socialInfra.NewCachedRepository(socialRepository, redisClient, cfg.Cache.PublicProfileTTL)
	}
	socialService := social.NewService(socialRepository)
	socialHandler := social.NewHandler(
		socialService,
		authService,
		social.WithNotifications(notificationHub),
		social.WithReadMiddlewares(readRateLimit),
	)
	postgresPostRepository := postInfra.NewPostgresRepository(db)
	var postRepository post.Repository = postgresPostRepository
	commentEventBroker := post.NewCommentEventBroker()
	postEventBroker := post.NewPostEventBroker()
	var commentEventPublisher post.CommentEventPublisher = commentEventBroker
	var postEventPublisher post.PostEventPublisher = postEventBroker
	var stopEngagementWorker context.CancelFunc
	var stopCommentEventSubscriber context.CancelFunc
	var stopPostEventSubscriber context.CancelFunc
	if redisClient != nil {
		postRepository = postInfra.NewEngagementStreamRepository(postRepository, redisClient)
		engagementWorkerCtx, cancelEngagementWorker := context.WithCancel(context.Background())
		stopEngagementWorker = cancelEngagementWorker
		go postInfra.RunEngagementStreamWorker(engagementWorkerCtx, postgresPostRepository, redisClient, cfg.Engagement)
		commentEventPublisher = notificationInfra.NewRedisCommentEventPublisher(redisClient, commentEventBroker)
		postEventPublisher = notificationInfra.NewRedisPostEventPublisher(redisClient, postEventBroker)
		commentEventCtx, cancelCommentEventSubscriber := context.WithCancel(context.Background())
		stopCommentEventSubscriber = cancelCommentEventSubscriber
		go notificationInfra.RunRedisCommentEventSubscriber(commentEventCtx, commentEventBroker, redisClient)
		postEventCtx, cancelPostEventSubscriber := context.WithCancel(context.Background())
		stopPostEventSubscriber = cancelPostEventSubscriber
		go notificationInfra.RunRedisPostEventSubscriber(postEventCtx, postEventBroker, redisClient)
	}
	var profileCacheInvalidator publicProfileCacheInvalidator
	if invalidator, ok := socialRepository.(publicProfileCacheInvalidator); ok {
		profileCacheInvalidator = invalidator
	}
	authService.SetAvatarEventPublisher(authAvatarEventPublisher{
		posts:    postEventPublisher,
		profiles: profileCacheInvalidator,
	})
	if redisClient != nil {
		postRepository = postInfra.NewCachedPostRepository(postRepository, redisClient, cfg.Cache.FeedFirstPageTTL)
	}
	postService := post.NewService(postRepository)
	commentRateLimit := httpMiddleware.NewKeyedRateLimiter(cfg.Spam.CommentRateLimitPerMin, time.Minute, userIDRateLimitKey)
	reportRateLimit := httpMiddleware.NewKeyedRateLimiter(cfg.Spam.ReportRateLimitPerHour, time.Hour, userIDRateLimitKey)
	postHandler := post.NewHandler(
		postService,
		authService,
		post.WithCommentEvents(commentEventBroker, commentEventPublisher),
		post.WithPostEvents(postEventBroker, postEventPublisher),
		post.WithNotifications(notificationHub),
		post.WithFollowers(socialService),
		post.WithReadMiddlewares(readRateLimit),
		post.WithCommentMiddlewares(commentRateLimit),
		post.WithReportMiddlewares(reportRateLimit),
	)
	imageRepository := uploadInfra.NewPostgresRepository(db)
	imageStorage := uploadInfra.NewSeaweedFSStorage(cfg.Upload)
	uploadService := upload.NewService(
		imageRepository,
		imageStorage,
		upload.WithMaxSize(cfg.Upload.MaxSize),
		upload.WithAllowedImageTypes(cfg.Upload.AllowedTypes),
	)
	uploadRateLimit := httpMiddleware.NewKeyedRateLimiter(cfg.Upload.RateLimitPerMin, time.Minute, userIDRateLimitKey)
	uploadHandler := upload.NewHandler(
		uploadService,
		authService,
		upload.WithProtectedMiddlewares(uploadRateLimit),
		upload.WithMaxBodyBytes(cfg.Upload.MaxSize+(1<<20)),
		upload.WithNotifications(notificationHub),
	)
	var categoryRepository category.Repository = categoryInfra.NewPostgresRepository(db)
	if redisClient != nil {
		categoryRepository = categoryInfra.NewCachedRepository(categoryRepository, redisClient, cfg.Cache.CategoriesTTL)
	}
	categoryService := category.NewService(categoryRepository)
	categoryHandler := category.NewHandler(categoryService, authService, category.WithReadMiddlewares(readRateLimit))
	r := chi.NewRouter()
	if cfg.HTTP.TrustProxyHeaders {
		r.Use(middleware.RealIP)
	}
	r.Use(httpMiddleware.CORS(httpMiddleware.CORSConfig{
		AllowedOrigins:   cfg.HTTP.CORSAllowedOrigins,
		AllowedMethods:   cfg.HTTP.CORSAllowedMethods,
		AllowedHeaders:   cfg.HTTP.CORSAllowedHeaders,
		AllowCredentials: cfg.HTTP.CORSAllowCredentials,
		MaxAgeSeconds:    cfg.HTTP.CORSMaxAgeSeconds,
	}))
	r.Use(middleware.RequestID)
	r.Use(httpMiddleware.Recover)
	r.Use(logger.RequestLogger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
	r.Route("/api", func(api chi.Router) {
		api.Use(httpMiddleware.InternalHeader(httpMiddleware.InternalHeaderConfig{
			Name:  cfg.HTTP.InternalHeaderName,
			Value: cfg.HTTP.InternalHeaderValue,
		}))
		api.Mount("/auth", authHandler.Routes())
		api.Mount("/locations", locationHandler.Routes())
		api.Mount("/posts", postHandler.Routes())
		api.Mount("/notifications", notificationHandler.Routes())
		api.Mount("/categories", categoryHandler.Routes())
		api.Mount("/users", socialHandler.Routes())
		api.Mount("/", uploadHandler.Routes())
	})

	sm := shutdown.NewManager()
	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           r,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}
	grpcListener, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", cfg.GRPC.Addr).Msg("failed to listen grpc")
	}
	sm.Register("auth-session-cleanup-stop", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
		stopSessionCleanup()
		return nil
	})
	if stopEngagementWorker != nil {
		sm.Register("post-engagement-worker-stop", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
			stopEngagementWorker()
			return nil
		})
	}
	if stopCommentEventSubscriber != nil {
		sm.Register("post-comment-event-subscriber-stop", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
			stopCommentEventSubscriber()
			return nil
		})
	}
	if stopPostEventSubscriber != nil {
		sm.Register("post-event-subscriber-stop", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
			stopPostEventSubscriber()
			return nil
		})
	}
	sm.Register("http-stop-accepting", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
		// Stop accepting new requests immediately
		return srv.Shutdown(ctx)
	})
	sm.Register("grpc-stop-accepting", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-ctx.Done():
			grpcServer.Stop()
			return ctx.Err()
		}
	})
	sm.Register("postgres-close", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
		return db.Close()
	})
	if redisClient != nil {
		sm.Register("redis-close", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
			return redisClient.Close()
		})
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error().Err(err).Str("addr", cfg.GRPC.Addr).Msg("grpc server failed")
			os.Exit(1)
		}
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Str("addr", srv.Addr).Msg("server failed")
			os.Exit(1)
		}
	}()
	log.Info().Str("addr", srv.Addr).Msg("server started")
	log.Info().Str("addr", cfg.GRPC.Addr).Msg("grpc server started")
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutdown initiated")

	if err := sm.Shutdown(); err != nil {
		log.Error().Err(err).Msg("shutdown completed with errors")
		os.Exit(1)
	}

	log.Info().Msg("shutdown complete")
}

func userIDRateLimitKey(r *http.Request) string {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		return ""
	}

	return strconv.FormatUint(principal.UserID, 10)
}
