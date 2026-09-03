//go:build wireinject

package main

import (
	"be/internal/api"
	"be/internal/api/handler"
	apimiddleware "be/internal/api/middleware"
	authapp "be/internal/application/auth"
	chatapp "be/internal/application/chat"
	roomapp "be/internal/application/room"
	socialapp "be/internal/application/social"
	"be/internal/config"
	"context"

	"github.com/goforj/wire"
)

func initializeApplication(ctx context.Context, cfg *config.Config) (*application, func(), error) {
	wire.Build(
		provideMetrics,
		provideLogger,
		provideDatabase,
		provideRedis,
		provideClock,
		provideUserRepository,
		provideRoomRepository,
		provideChatRepository,
		provideSocialRepository,
		providePasswordHasher,
		provideTokenManager,
		provideTokenSessionStore,
		provideGoogleIdentityVerifier,
		provideInviteCodeGenerator,
		authapp.NewRegisterUseCase,
		provideLoginUseCase,
		authapp.NewRefreshTokenUseCase,
		authapp.NewForgotPasswordUseCase,
		authapp.NewResetPasswordUseCase,
		authapp.NewLogoutUseCase,
		authapp.NewGoogleLoginUseCase,
		provideCreateRoomUseCase,
		roomapp.NewListRoomsUseCase,
		roomapp.NewGetRoomUseCase,
		roomapp.NewJoinRoomUseCase,
		roomapp.NewKickMemberUseCase,
		roomapp.NewDealRoundUseCase,
		roomapp.NewGetCurrentCardUseCase,
		roomapp.NewUpdateDiscussionUseCase,
		roomapp.NewGetRoundStateUseCase,
		roomapp.NewCastVoteUseCase,
		roomapp.NewPlayerReadyUseCase,
		roomapp.NewFinishTurnUseCase,
		roomapp.NewMrWhiteGuessUseCase,
		chatapp.NewService,
		socialapp.NewService,
		provideRealtimeBackplane,
		provideRealtimeHub,
		provideAuthHandler,
		handler.NewRoomHandler,
		provideRealtimeHandler,
		handler.NewChatHandler,
		handler.NewSocialHandler,
		providePhaseTransitionHandler,
		roomapp.NewPhaseScheduler,
		provideHealthHandler,
		apimiddleware.NewAuthenticator,
		provideMetricsHandler,
		api.NewRouter,
		provideHTTPServer,
		wire.Struct(new(application), "*"),
	)
	return nil, nil, nil
}
