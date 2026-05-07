package category

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, name, slug string) error {
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
	return s.repository.GetByID(ctx, id)
}
func (s *Service) GetBySlug(ctx context.Context, slug string) (Category, error) {
	return s.repository.GetBySlug(ctx, slug)
}
func (s *Service) List(ctx context.Context) ([]Category, error) {
	return s.repository.List(ctx)
}
func (s *Service) Update(ctx context.Context, id uint64, name, slug string) (Category, error) {
	category, err := s.GetByID(ctx, id)
	if err != nil {
		return Category{}, err
	}
	category.Name = name
	category.Slug = slug
	return s.repository.Update(ctx, id, name, slug)
}
func (s *Service) Delete(ctx context.Context, id uint64) error {
	return s.repository.Delete(ctx, id)
}
