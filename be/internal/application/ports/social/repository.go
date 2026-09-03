package social

import (
	domainsocial "be/internal/domain/social"
	"context"
	"time"
)

type Repository interface {
	SearchUsers(ctx context.Context, currentUserID int64, query string, limit int) ([]domainsocial.UserSummary, error)
	CreateFriendRequest(ctx context.Context, senderID, receiverID int64, now time.Time) (*domainsocial.FriendRequest, error)
	ListPendingFriendRequests(ctx context.Context, userID int64) ([]domainsocial.FriendRequest, error)
	RespondFriendRequest(ctx context.Context, userID, requestID int64, accept bool, now time.Time) (*domainsocial.FriendRequest, error)
	CancelFriendRequest(ctx context.Context, userID, requestID int64, now time.Time) error
	ListFriends(ctx context.Context, userID int64) ([]domainsocial.Friend, error)
	Unfriend(ctx context.Context, userID, friendUserID int64) error
	ListNotifications(ctx context.Context, userID int64, unreadOnly bool, limit, offset int) ([]domainsocial.Notification, error)
	CountUnreadNotifications(ctx context.Context, userID int64) (int, error)
	MarkNotificationRead(ctx context.Context, userID, notificationID int64, now time.Time) error
	MarkAllNotificationsRead(ctx context.Context, userID int64, now time.Time) (int64, error)
}
