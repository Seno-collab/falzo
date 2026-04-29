package infrastructure

import (
	"context"

	"falzo-be/internal/location/domain"
	"falzo-be/internal/location/domain/entity"
	"falzo-be/internal/location/domain/repository"
	"falzo-be/pkg/database"
)

const locationRepoService = "location"

type LocationRepositoryPG struct {
	db database.Client
}

func NewLocationRepositoryPG(db database.Client) repository.LocationRepository {
	return &LocationRepositoryPG{db: db}
}

func (r *LocationRepositoryPG) Search(ctx context.Context, query string) ([]entity.Location, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, domain.ErrLocationDependencyUnavailable
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

	locations := make([]entity.Location, 0)
	for rows.Next() {
		var location entity.Location
		if err := rows.Scan(
			&location.ID,
			&location.Name,
			&location.Address,
			&location.Latitude,
			&location.Longitude,
		); err != nil {
			return nil, mapDBError(ctx, locationRepoService, "locations.search.scan", err)
		}

		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.search.iterate", err)
	}

	return locations, nil
}

func (r *LocationRepositoryPG) Nearby(ctx context.Context, latitude, longitude, radiusMeters float64) ([]entity.NearbyLocation, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, domain.ErrLocationDependencyUnavailable
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

	locations := make([]entity.NearbyLocation, 0)
	for rows.Next() {
		var location entity.NearbyLocation
		if err := rows.Scan(
			&location.Location.ID,
			&location.Location.Name,
			&location.Location.Address,
			&location.Location.Latitude,
			&location.Location.Longitude,
			&location.DistanceMeters,
		); err != nil {
			return nil, mapDBError(ctx, locationRepoService, "locations.nearby.scan", err)
		}

		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.nearby.iterate", err)
	}

	return locations, nil
}

func (r *LocationRepositoryPG) GetPostsByLocationID(ctx context.Context, locationID string) ([]entity.LocationPost, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, domain.ErrLocationDependencyUnavailable
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

	posts := make([]entity.LocationPost, 0)
	for rows.Next() {
		var post entity.LocationPost
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.ImageURL,
			&post.Caption,
			&post.LocationName,
			&post.Latitude,
			&post.Longitude,
		); err != nil {
			return nil, mapDBError(ctx, locationRepoService, "locations.get_posts_by_location.scan", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.get_posts_by_location.iterate", err)
	}

	return posts, nil
}
