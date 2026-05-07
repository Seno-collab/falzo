package category

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, name, slug string) error {
	if s.repository == nil {
		return ErrDependencyUnavailable
	}

	nameObj, err := NewName(name)
	if err != nil {
		return err
	}
	slugObj, err := NewSlug(slug)
	if err != nil {
		return err
	}

	input := CategoryCreateInput{
		Name: nameObj.String(),
		Slug: slugObj.String(),
	}
	return s.repository.Create(ctx, input)
}

func (s *Service) GetByID(ctx context.Context, id uint64) (Category, error) {
	if s.repository == nil {
		return Category{}, ErrDependencyUnavailable
	}

	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Category, error) {
	if s.repository == nil {
		return Category{}, ErrDependencyUnavailable
	}

	slugObj, err := NewSlug(slug)
	if err != nil {
		return Category{}, err
	}

	return s.repository.GetBySlug(ctx, slugObj.String())
}

func (s *Service) List(ctx context.Context) ([]Category, error) {
	if s.repository == nil {
		return nil, ErrDependencyUnavailable
	}

	return s.repository.List(ctx)
}

func (s *Service) Update(ctx context.Context, id uint64, name, slug string) (Category, error) {
	if s.repository == nil {
		return Category{}, ErrDependencyUnavailable
	}

	nameObj, err := NewName(name)
	if err != nil {
		return Category{}, err
	}
	slugObj, err := NewSlug(slug)
	if err != nil {
		return Category{}, err
	}

	return s.repository.Update(ctx, id, nameObj.String(), slugObj.String())
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	if s.repository == nil {
		return ErrDependencyUnavailable
	}

	return s.repository.Delete(ctx, id)
}
