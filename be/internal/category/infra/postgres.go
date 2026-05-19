package categoryInfra

import (
	"context"
	"errors"
	"falzo-be/internal/category"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
	"falzo-be/pkg/database/orm"
	"falzo-be/pkg/dberr"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	db         database.Client
	categories *orm.Table[category.Category]
}

const categoryRepoService = "category"

func NewPostgresRepository(db database.Client) *PostgresRepository {
	repository := &PostgresRepository{db: db}
	if db != nil && db.Pool() != nil {
		repository.categories = newCategoryTable(db.Pool())
	}
	return repository
}

func (r *PostgresRepository) Create(ctx context.Context, input category.CategoryCreateInput) error {
	table, err := r.table()
	if err != nil {
		return err
	}
	_, err = table.Insert(ctx, orm.Values{
		"name": input.Name,
		"slug": input.Slug,
	})
	if err != nil {
		if dberr.IsUniqueViolation(err) {
			return category.ErrAlreadyExists
		}
		return mapDBError(ctx, categoryRepoService, "categories.create", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uint64) (category.Category, error) {
	table, err := r.table()
	if err != nil {
		return category.Category{}, err
	}
	cat, err := table.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category.Category{}, category.ErrNotFound
		}
		return category.Category{}, mapDBError(ctx, categoryRepoService, "categories.getByID", err)
	}
	return cat, nil
}

func (r *PostgresRepository) GetBySlug(ctx context.Context, slug string) (category.Category, error) {
	table, err := r.table()
	if err != nil {
		return category.Category{}, err
	}
	cat, err := table.FindOne(ctx, "slug = $1", slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category.Category{}, category.ErrNotFound
		}
		return category.Category{}, mapDBError(ctx, categoryRepoService, "categories.getBySlug", err)
	}
	return cat, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]category.Category, error) {
	table, err := r.table()
	if err != nil {
		return []category.Category{}, err
	}
	categories, err := table.List(ctx, orm.QueryOptions{})
	if err != nil {
		return []category.Category{}, mapDBError(ctx, categoryRepoService, "categories.list.iterate", err)
	}
	return categories, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id uint64, name, slug string) (category.Category, error) {
	table, err := r.table()
	if err != nil {
		return category.Category{}, err
	}
	result, err := table.UpdateByID(ctx, id, orm.Values{
		"name": name,
		"slug": slug,
	})
	if err != nil {
		if dberr.IsUniqueViolation(err) {
			return category.Category{}, category.ErrAlreadyExists
		}
		return category.Category{}, mapDBError(ctx, categoryRepoService, "categories.update", err)
	}
	if result.RowsAffected() == 0 {
		return category.Category{}, category.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *PostgresRepository) Delete(ctx context.Context, id uint64) error {
	table, err := r.table()
	if err != nil {
		return err
	}
	result, err := table.DeleteByID(ctx, id)
	if err != nil {
		return mapDBError(ctx, categoryRepoService, "categories.delete", err)
	}
	if result.RowsAffected() == 0 {
		return category.ErrNotFound
	}
	return nil
}

func mapDBError(ctx context.Context, service, operation string, err error) error {
	return share.MapDBError(ctx, service, operation, err, category.ErrDependencyUnavailable, category.ErrInternal)
}

func (r *PostgresRepository) table() (*orm.Table[category.Category], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, category.ErrDependencyUnavailable
	}
	if r.categories != nil {
		return r.categories, nil
	}
	return newCategoryTable(r.db.Pool()), nil
}

func newCategoryTable(db orm.Queryer) *orm.Table[category.Category] {
	return orm.NewTable(db, "categories", []string{"id", "name", "slug"}, scanCategory)
}

func scanCategory(scanner orm.Scanner) (category.Category, error) {
	var cat category.Category
	err := scanner.Scan(&cat.ID, &cat.Name, &cat.Slug)
	return cat, err
}
