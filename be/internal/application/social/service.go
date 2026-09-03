package social

import (
	socialports "be/internal/application/ports/social"
	domainsocial "be/internal/domain/social"
	"be/internal/shared/clock"
	"context"
	"strings"
)

const (
	DefaultNotificationLimit = 30
	MaxNotificationLimit     = 100
	DefaultUserSearchLimit   = 20
	MaxUserSearchLimit       = 50
)

type Service struct {
	repository socialports.Repository
	clock      clock.Clock
}

func NewService(repository socialports.Repository, clock clock.Clock) *Service {
	return &Service{repository: repository, clock: clock}
}

func (s *Service) SearchUsers(ctx context.Context, currentUserID int64, query string, limit int) ([]domainsocial.UserSummary, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return nil, domainsocial.ErrSearchQueryTooShort
	}
	if limit <= 0 {
		limit = DefaultUserSearchLimit
	}
	if limit > MaxUserSearchLimit {
		limit = MaxUserSearchLimit
	}
	return s.repository.SearchUsers(ctx, currentUserID, query, limit)
}

func (s *Service) SendFriendRequest(ctx context.Context, senderID, receiverID int64) (*domainsocial.FriendRequest, error) {
	if senderID == receiverID {
		return nil, domainsocial.ErrCannotFriendSelf
	}
	return s.repository.CreateFriendRequest(ctx, senderID, receiverID, s.clock.Now())
}

func (s *Service) ListPendingFriendRequests(ctx context.Context, userID int64) ([]domainsocial.FriendRequest, error) {
	return s.repository.ListPendingFriendRequests(ctx, userID)
}

func (s *Service) AcceptFriendRequest(ctx context.Context, userID, requestID int64) (*domainsocial.FriendRequest, error) {
	return s.repository.RespondFriendRequest(ctx, userID, requestID, true, s.clock.Now())
}

func (s *Service) RejectFriendRequest(ctx context.Context, userID, requestID int64) (*domainsocial.FriendRequest, error) {
	return s.repository.RespondFriendRequest(ctx, userID, requestID, false, s.clock.Now())
}

func (s *Service) CancelFriendRequest(ctx context.Context, userID, requestID int64) error {
	return s.repository.CancelFriendRequest(ctx, userID, requestID, s.clock.Now())
}

func (s *Service) ListFriends(ctx context.Context, userID int64) ([]domainsocial.Friend, error) {
	return s.repository.ListFriends(ctx, userID)
}

func (s *Service) Unfriend(ctx context.Context, userID, friendUserID int64) error {
	if userID == friendUserID {
		return domainsocial.ErrFriendshipNotFound
	}
	return s.repository.Unfriend(ctx, userID, friendUserID)
}

func (s *Service) ListNotifications(ctx context.Context, userID int64, unreadOnly bool, limit, offset int) ([]domainsocial.Notification, error) {
	if limit <= 0 {
		limit = DefaultNotificationLimit
	}
	if limit > MaxNotificationLimit {
		limit = MaxNotificationLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repository.ListNotifications(ctx, userID, unreadOnly, limit, offset)
}

func (s *Service) CountUnreadNotifications(ctx context.Context, userID int64) (int, error) {
	return s.repository.CountUnreadNotifications(ctx, userID)
}

func (s *Service) MarkNotificationRead(ctx context.Context, userID, notificationID int64) error {
	return s.repository.MarkNotificationRead(ctx, userID, notificationID, s.clock.Now())
}

func (s *Service) MarkAllNotificationsRead(ctx context.Context, userID int64) (int64, error) {
	return s.repository.MarkAllNotificationsRead(ctx, userID, s.clock.Now())
}
