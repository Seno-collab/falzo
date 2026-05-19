package infra

import (
	"context"

	"falzo-be/internal/location"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
	"falzo-be/pkg/database/orm"
)

const locationRepoService = "location"

type PostgresRepository struct {
	db              database.Client
	locations       *orm.Table[location.Location]
	nearbyLocations *orm.Table[location.NearbyLocation]
	locationPosts   *orm.Table[location.LocationPost]
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	repository := &PostgresRepository{db: db}
	if db != nil && db.Pool() != nil {
		repository.locations = newLocationTable(db.Pool())
		repository.nearbyLocations = newNearbyLocationTable(db.Pool())
		repository.locationPosts = newLocationPostTable(db.Pool())
	}
	return repository
}

func (r *PostgresRepository) Search(ctx context.Context, query string) ([]location.Location, error) {
	table, err := r.locationTable()
	if err != nil {
		return nil, err
	}

	locations, err := table.List(ctx, orm.QueryOptions{
		Where:   "name ILIKE '%' || $1 || '%'",
		Args:    []any{query},
		OrderBy: `"name" ASC`,
		Limit:   50,
	})
	if err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.search", err)
	}

	return locations, nil
}

func (r *PostgresRepository) Nearby(ctx context.Context, latitude, longitude, radiusMeters float64) ([]location.NearbyLocation, error) {
	table, err := r.nearbyLocationTable()
	if err != nil {
		return nil, err
	}

	locations, err := table.List(ctx, orm.QueryOptions{
		Where: `ST_DWithin(
			geom,
			ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
			$3
		)`,
		Args:    []any{latitude, longitude, radiusMeters},
		OrderBy: "distance_meters ASC",
		Limit:   100,
	})
	if err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.nearby", err)
	}

	return locations, nil
}

func (r *PostgresRepository) GetPostsByLocationID(ctx context.Context, locationID string) ([]location.LocationPost, error) {
	table, err := r.locationPostTable()
	if err != nil {
		return nil, err
	}

	posts, err := table.List(ctx, orm.QueryOptions{
		Where: "p.location_id::text = $1",
		Args:  []any{locationID},
	})
	if err != nil {
		return nil, mapDBError(ctx, locationRepoService, "locations.get_posts_by_location", err)
	}

	return posts, nil
}

func mapDBError(ctx context.Context, service, operation string, err error) error {
	return share.MapDBError(ctx, service, operation, err, location.ErrDependencyUnavailable, location.ErrInternal)
}

func (r *PostgresRepository) locationTable() (*orm.Table[location.Location], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, location.ErrDependencyUnavailable
	}
	if r.locations != nil {
		return r.locations, nil
	}
	return newLocationTable(r.db.Pool()), nil
}

func (r *PostgresRepository) nearbyLocationTable() (*orm.Table[location.NearbyLocation], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, location.ErrDependencyUnavailable
	}
	if r.nearbyLocations != nil {
		return r.nearbyLocations, nil
	}
	return newNearbyLocationTable(r.db.Pool()), nil
}

func (r *PostgresRepository) locationPostTable() (*orm.Table[location.LocationPost], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, location.ErrDependencyUnavailable
	}
	if r.locationPosts != nil {
		return r.locationPosts, nil
	}
	return newLocationPostTable(r.db.Pool()), nil
}

func newLocationTable(db orm.Queryer) *orm.Table[location.Location] {
	return orm.NewTable(db, "locations", []string{"id::text", "name", "address", "latitude", "longitude"}, scanLocation)
}

func newNearbyLocationTable(db orm.Queryer) *orm.Table[location.NearbyLocation] {
	return orm.NewTable(
		db,
		"locations",
		[]string{
			"id::text",
			"name",
			"address",
			"latitude",
			"longitude",
			"ST_Distance(geom, ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography) AS distance_meters",
		},
		scanNearbyLocation,
	)
}

func newLocationPostTable(db orm.Queryer) *orm.Table[location.LocationPost] {
	return orm.NewTable(
		db,
		"posts p",
		[]string{
			"p.id::text",
			"p.user_id",
			"p.image_url",
			"COALESCE(p.caption, '')",
			"COALESCE(p.location_name, '')",
			"COALESCE(p.latitude, 0)",
			"COALESCE(p.longitude, 0)",
		},
		scanLocationPost,
	)
}

func scanLocation(scanner orm.Scanner) (location.Location, error) {
	var item location.Location
	err := scanner.Scan(&item.ID, &item.Name, &item.Address, &item.Latitude, &item.Longitude)
	return item, err
}

func scanNearbyLocation(scanner orm.Scanner) (location.NearbyLocation, error) {
	var item location.NearbyLocation
	err := scanner.Scan(
		&item.Location.ID,
		&item.Location.Name,
		&item.Location.Address,
		&item.Location.Latitude,
		&item.Location.Longitude,
		&item.DistanceMeters,
	)
	return item, err
}

func scanLocationPost(scanner orm.Scanner) (location.LocationPost, error) {
	var item location.LocationPost
	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.ImageURL,
		&item.Caption,
		&item.LocationName,
		&item.Latitude,
		&item.Longitude,
	)
	return item, err
}

var _ location.Repository = (*PostgresRepository)(nil)
