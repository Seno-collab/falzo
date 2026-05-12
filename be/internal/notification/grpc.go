package notification

import (
	"context"
	"strconv"
	"strings"

	"falzo-be/internal/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

const serviceName = "falzo.notification.v1.NotificationService"

type Authenticator interface {
	Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
}

type GRPCServer struct {
	subscriber Subscriber
	auth       Authenticator
}

func NewGRPCServer(subscriber Subscriber, authService Authenticator) *GRPCServer {
	return &GRPCServer{subscriber: subscriber, auth: authService}
}

type NotificationServiceServer interface {
	Subscribe(*emptypb.Empty, NotificationService_SubscribeServer) error
}

type NotificationService_SubscribeServer interface {
	Send(*structpb.Struct) error
	grpc.ServerStream
}

func RegisterNotificationServiceServer(server *grpc.Server, service NotificationServiceServer) {
	server.RegisterService(&NotificationService_ServiceDesc, service)
}

func (s *GRPCServer) Subscribe(_ *emptypb.Empty, stream NotificationService_SubscribeServer) error {
	if s == nil || s.subscriber == nil {
		return status.Error(codes.Unavailable, "notification service unavailable")
	}

	principal, err := s.authenticate(stream.Context())
	if err != nil {
		return err
	}

	notifications, unsubscribe, err := s.subscriber.Subscribe(stream.Context(), principal.UserID)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	defer unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case item, ok := <-notifications:
			if !ok {
				return nil
			}

			payload, err := notificationStruct(item)
			if err != nil {
				return status.Error(codes.Internal, "notification payload encode failed")
			}
			if err := stream.Send(payload); err != nil {
				return err
			}
		}
	}
}

func (s *GRPCServer) authenticate(ctx context.Context) (*auth.AuthenticatedUser, error) {
	if s == nil || s.auth == nil {
		return nil, status.Error(codes.Unauthenticated, "auth service unavailable")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authorization metadata is required")
	}

	authHeader := ""
	for _, key := range []string{"authorization", "Authorization"} {
		values := md.Get(key)
		if len(values) > 0 {
			authHeader = strings.TrimSpace(values[0])
			break
		}
	}

	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return nil, status.Error(codes.Unauthenticated, "bearer token is required")
	}

	principal, err := s.auth.Authenticate(ctx, strings.TrimSpace(token))
	if err != nil || principal == nil || principal.UserID == 0 {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return principal, nil
}

func notificationStruct(item Notification) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{
		"id":            item.ID,
		"user_id":       item.UserID,
		"actor_user_id": item.ActorUserID,
		"actor_name":    item.ActorName,
		"type":          item.Type,
		"title":         item.Title,
		"body":          item.Body,
		"resource":      item.Resource,
		"resource_id":   item.ResourceID,
		"post_id":       item.PostID,
		"image_id":      item.ImageID,
		"created_at":    item.CreatedAt.UTC().Format(timeFormatRFC3339Nano),
	})
}

const timeFormatRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"

var NotificationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: serviceName,
	HandlerType: (*NotificationServiceServer)(nil),
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _NotificationService_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/notification/v1/notification.proto",
}

func _NotificationService_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	request := new(emptypb.Empty)
	if err := stream.RecvMsg(request); err != nil {
		return err
	}
	return srv.(NotificationServiceServer).Subscribe(request, &notificationServiceSubscribeServer{ServerStream: stream})
}

type notificationServiceSubscribeServer struct {
	grpc.ServerStream
}

func (s *notificationServiceSubscribeServer) Send(item *structpb.Struct) error {
	return s.ServerStream.SendMsg(item)
}

func ResourceIDUint64(id uint64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatUint(id, 10)
}

func ResourceIDInt64(id int64) string {
	if id <= 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
