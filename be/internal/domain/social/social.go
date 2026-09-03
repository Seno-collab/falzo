package social

import (
	"errors"
	"time"
)

type FriendRequestStatus string

const (
	FriendRequestPending  FriendRequestStatus = "PENDING"
	FriendRequestAccepted FriendRequestStatus = "ACCEPTED"
	FriendRequestRejected FriendRequestStatus = "REJECTED"
	FriendRequestCanceled FriendRequestStatus = "CANCELED"
)

type RelationshipStatus string

const (
	RelationshipNone            RelationshipStatus = "NONE"
	RelationshipFriends         RelationshipStatus = "FRIENDS"
	RelationshipIncomingRequest RelationshipStatus = "INCOMING_REQUEST"
	RelationshipOutgoingRequest RelationshipStatus = "OUTGOING_REQUEST"
)

type NotificationType string

const (
	NotificationFriendRequestReceived NotificationType = "FRIEND_REQUEST_RECEIVED"
	NotificationFriendRequestAccepted NotificationType = "FRIEND_REQUEST_ACCEPTED"
)

type UserSummary struct {
	ID           int64
	UserName     string
	Relationship RelationshipStatus
}

type FriendRequest struct {
	ID           int64
	SenderID     int64
	SenderName   string
	ReceiverID   int64
	ReceiverName string
	Status       FriendRequestStatus
	CreatedAt    time.Time
	RespondedAt  *time.Time
}

type Friend struct {
	ID        int64
	UserName  string
	FriendsAt time.Time
}

type Notification struct {
	ID          int64
	Type        NotificationType
	ActorID     int64
	ActorName   string
	ReferenceID int64
	ReadAt      *time.Time
	CreatedAt   time.Time
}

var (
	ErrCannotFriendSelf        = errors.New("cannot send a friend request to yourself")
	ErrSearchQueryTooShort     = errors.New("user search query must contain at least two characters")
	ErrUserNotFound            = errors.New("social user not found")
	ErrAlreadyFriends          = errors.New("users are already friends")
	ErrFriendRequestExists     = errors.New("a pending friend request already exists")
	ErrFriendRequestNotFound   = errors.New("friend request not found")
	ErrFriendRequestNotPending = errors.New("friend request is not pending")
	ErrFriendRequestForbidden  = errors.New("friend request operation is forbidden")
	ErrFriendshipNotFound      = errors.New("friendship not found")
	ErrNotificationNotFound    = errors.New("notification not found")
)
