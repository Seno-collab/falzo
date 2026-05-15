package social

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

type PublicProfileInput struct {
	UserID       uint64
	ViewerUserID uint64
}

type FollowInput struct {
	FollowerID  uint64
	FollowingID uint64
}

func (s *Service) GetPublicProfile(ctx context.Context, input PublicProfileInput) (PublicProfile, error) {
	if s.repository == nil {
		return PublicProfile{}, ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return PublicProfile{}, ErrUserIDRequired
	}

	return s.repository.GetPublicProfile(ctx, input.UserID, input.ViewerUserID)
}

func (s *Service) Follow(ctx context.Context, input FollowInput) (bool, error) {
	if s.repository == nil {
		return false, ErrDependencyUnavailable
	}
	if input.FollowerID == 0 {
		return false, ErrUserIDRequired
	}
	if input.FollowingID == 0 {
		return false, ErrTargetUserIDRequired
	}
	if input.FollowerID == input.FollowingID {
		return false, ErrCannotFollowSelf
	}

	return s.repository.Follow(ctx, input.FollowerID, input.FollowingID)
}

func (s *Service) Unfollow(ctx context.Context, input FollowInput) error {
	if s.repository == nil {
		return ErrDependencyUnavailable
	}
	if input.FollowerID == 0 {
		return ErrUserIDRequired
	}
	if input.FollowingID == 0 {
		return ErrTargetUserIDRequired
	}
	if input.FollowerID == input.FollowingID {
		return ErrCannotFollowSelf
	}

	return s.repository.Unfollow(ctx, input.FollowerID, input.FollowingID)
}

func (s *Service) Block(ctx context.Context, input FollowInput) error {
	if s.repository == nil {
		return ErrDependencyUnavailable
	}
	if input.FollowerID == 0 {
		return ErrUserIDRequired
	}
	if input.FollowingID == 0 {
		return ErrTargetUserIDRequired
	}
	if input.FollowerID == input.FollowingID {
		return ErrCannotBlockSelf
	}

	return s.repository.Block(ctx, input.FollowerID, input.FollowingID)
}

func (s *Service) Unblock(ctx context.Context, input FollowInput) error {
	if s.repository == nil {
		return ErrDependencyUnavailable
	}
	if input.FollowerID == 0 {
		return ErrUserIDRequired
	}
	if input.FollowingID == 0 {
		return ErrTargetUserIDRequired
	}
	if input.FollowerID == input.FollowingID {
		return ErrCannotBlockSelf
	}

	return s.repository.Unblock(ctx, input.FollowerID, input.FollowingID)
}

func (s *Service) ListFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	if s.repository == nil {
		return nil, ErrDependencyUnavailable
	}
	if userID == 0 {
		return nil, ErrUserIDRequired
	}

	return s.repository.ListFollowerIDs(ctx, userID)
}
