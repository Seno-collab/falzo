package infra

import (
	"context"

	"falzo-be/internal/location"
	"falzo-be/pkg/database"
	"falzo-be/pkg/dberr"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

const locationRepoService = "location"

type PostgresRepository struct {
	db database.Client
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Search(ctx context.Context, query string) ([]location.Location, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, location.ErrDependencyUnavailable
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT id::text, name, address, latitude, longitude
		FROM locations
		WHERE name ILIKE '%' || $1 || '%'
		ORDER BY name ASC
		LIMIT 50
	`, query)
	if err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.search", err)
	}
	defer rows.Close()

	locations := make([]location.Location, 0)
	for rows.Next() {
		var item location.Location
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Address,
			&item.Latitude,
			&item.Longitude,
		); err != nil {
			return nil, mapDBError(ctx, locationRepoService, "locations.search.scan", err)
		}

		locations = append(locations, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.search.iterate", err)
	}

	return locations, nil
}

func (r *PostgresRepository) Nearby(ctx context.Context, latitude, longitude, radiusMeters float64) ([]location.NearbyLocation, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, location.ErrDependencyUnavailable
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT
			id::text,
			name,
			address,
			latitude,
			longitude,
			ST_Distance(
				geom,
				ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography
			) AS distance_meters
		FROM locations
		WHERE ST_DWithin(
			geom,
			ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
			$3
		)
		ORDER BY distance_meters ASC
		LIMIT 100
	`, latitude, longitude, radiusMeters)
	if err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.nearby", err)
	}
	defer rows.Close()

	locations := make([]location.NearbyLocation, 0)
	for rows.Next() {
		var item location.NearbyLocation
		if err := rows.Scan(
			&item.Location.ID,
			&item.Location.Name,
			&item.Location.Address,
			&item.Location.Latitude,
			&item.Location.Longitude,
			&item.DistanceMeters,
		); err != nil {
			return nil, mapDBError(ctx, locationRepoService, "locations.nearby.scan", err)
		}

		locations = append(locations, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.nearby.iterate", err)
	}

	return locations, nil
}

func (r *PostgresRepository) GetPostsByLocationID(ctx context.Context, locationID string) ([]location.LocationPost, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, location.ErrDependencyUnavailable
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT
			p.id::text,
			p.user_id,
			p.image_url,
			COALESCE(p.caption, ''),
			COALESCE(p.location_name, ''),
			COALESCE(p.latitude, 0),
			COALESCE(p.longitude, 0)
		FROM posts p
		WHERE p.location_id::text = $1
	`, locationID)
	if err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.get_posts_by_location", err)
	}
	defer rows.Close()

	posts := make([]location.LocationPost, 0)
	for rows.Next() {
		var item location.LocationPost
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ImageURL,
			&item.Caption,
			&item.LocationName,
			&item.Latitude,
			&item.Longitude,
		); err != nil {
			return nil, mapDBError(ctx, locationRepoService, "locations.get_posts_by_location.scan", err)
		}

		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.get_posts_by_location.iterate", err)
	}

	return posts, nil
}

func mapDBError(ctx context.Context, service, operation string, err error) error {
	return dberr.MapDependencyOrInternal(
		err,
		service,
		operation,
		chimiddleware.GetReqID(ctx),
		location.ErrDependencyUnavailable,
		location.ErrInternal,
	)
}

var _ location.Repository = (*PostgresRepository)(nil)
