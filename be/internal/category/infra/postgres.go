package categoryInfra

import (
	"context"
	"falzo-be/internal/category"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
)

type PostgresRepository struct {
	db database.Client
}

const categoryRepoService = "category"

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, input category.CategoryCreateInput) error {
	// Implementation for creating a category in the database
	if r.db == nil || r.db.Pool() == nil {
		return category.ErrDependencyUnavailable
	}
	_, err := r.db.Pool().Exec(ctx, `
		INSERT INTO categories (name, slug)
		VALUES ($1, $2)
	`, input.Name, input.Slug)
	if err != nil {
		return mapDBError(ctx, categoryRepoService, "categories.create", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint64) (category.Category, error) {
	// Implementation for getting a category by ID from the database
	if db := r.db; db == nil || db.Pool() == nil {
		return category.Category{}, category.ErrDependencyUnavailable
	}
	var cat category.Category
	err := r.db.Pool().QueryRow(ctx, `
		SELECT id, name, slug
		FROM categories
		WHERE id = $1
	`, id).Scan(&cat.ID, &cat.Name, &cat.Slug)
	if err != nil {
		return category.Category{}, mapDBError(ctx, categoryRepoService, "categories.getByID", err)
	}
	return cat, nil
}

func (r *PostgresRepository) GetBySlug(ctx context.Context, slug string) (category.Category, error) {
	// Implementation for getting a category by slug from the database
	if db := r.db; db == nil || db.Pool() == nil {
		return category.Category{}, category.ErrDependencyUnavailable
	}
	var cat category.Category
	err := r.db.Pool().QueryRow(ctx, `
		SELECT id, name, slug
		FROM categories
		WHERE slug = $1
	`, slug).Scan(&cat.ID, &cat.Name, &cat.Slug)
	if err != nil {
		return category.Category{}, mapDBError(ctx, categoryRepoService, "categories.getBySlug", err)
	}
	return cat, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]category.Category, error) {
	// Implementation for listing categories from the database
	if db := r.db; db == nil || db.Pool() == nil {
		return []category.Category{}, category.ErrDependencyUnavailable
	}
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, name, slug
		FROM categories
	`)
	if err != nil {
		return []category.Category{}, mapDBError(ctx, categoryRepoService, "categories.list", err)
	}
	defer rows.Close()

	var categories []category.Category
	for rows.Next() {
		var cat category.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug); err != nil {
			return []category.Category{}, mapDBError(ctx, categoryRepoService, "categories.list", err)
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uint64, name, slug string) (category.Category, error) {
	// Implementation for updating a category in the database
	if r.db == nil || r.db.Pool() == nil {
		return category.Category{}, category.ErrDependencyUnavailable
	}
	_, err := r.db.Pool().Exec(ctx, `
		UPDATE categories
		SET name = $1, slug = $2
		WHERE id = $3
	`, name, slug, id)
	if err != nil {
		return category.Category{}, mapDBError(ctx, categoryRepoService, "categories.update", err)
	}
	return r.GetByID(ctx, id)
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint64) error {
	// Implementation for deleting a category from the database
	if r.db == nil || r.db.Pool() == nil {
		return category.ErrDependencyUnavailable
	}
	_, err := r.db.Pool().Exec(ctx, `
		DELETE FROM categories
		WHERE id = $1
	`, id)
	if err != nil {
		return mapDBError(ctx, categoryRepoService, "categories.delete", err)
	}
	return nil
}

func mapDBError(ctx context.Context, service, operation string, err error) error {
	// Map database errors to application-specific errors
	return share.MapDBError(ctx, service, operation, err, category.ErrDependencyUnavailable, category.ErrInternal)
}
