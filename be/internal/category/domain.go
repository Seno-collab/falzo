package category

import (
	"context"
	"errors"
)

var (
	ErrDependencyUnavailable = errors.New("Category dependency unavailable")
	ErrInternal              = errors.New("Category internal error")
	ErrNameRequired          = errors.New("Category name is required")
	ErrSlugRequired          = errors.New("Category slug is required")
	ErrNameTooLong           = errors.New("Category name cannot exceed 255 characters")
	ErrSlugTooLong           = errors.New("Category slug cannot exceed 255 characters")
)

type Repository interface {
	Create(context context.Context, input CategoryCreateInput) error
	GetByID(context context.Context, id uint64) (Category, error)
	GetBySlug(context context.Context, slug string) (Category, error)
	List(context context.Context) ([]Category, error)
	Update(context context.Context, id uint64, name, slug string) (Category, error)
	Delete(context context.Context, id uint64) error
}

type Category struct {
	ID   uint64
	Name string
	Slug string
}

type CategoryCreateInput struct {
	Name string
	Slug string
}

type Name string

func NewCategory(name, slug string) (Category, error) {
	nameObj, err := NewName(name)
	if err != nil {
		return Category{}, err
	}
	slugObj, err := NewSlug(slug)
	if err != nil {
		return Category{}, err
	}
	return Category{
		Name: nameObj.String(),
		Slug: slugObj.String(),
	}, nil
}

func NewName(name string) (Name, error) {
	if name == "" {
		return Name(""), ErrNameRequired
	}
	if len(name) > 255 {
		return Name(""), ErrNameTooLong
	}
	return Name(name), nil
}

func (n Name) String() string {
	return string(n)
}

type Slug string

func NewSlug(slug string) (Slug, error) {
	if slug == "" {
		return Slug(""), ErrSlugRequired
	}
	if len(slug) > 255 {
		return Slug(""), ErrSlugTooLong
	}
	return Slug(slug), nil
}

func (s Slug) String() string {
	return string(s)
}
