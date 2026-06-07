package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"falzo-be/internal/itinerary"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
	"falzo-be/pkg/database/orm"

	"github.com/jackc/pgx/v5"
)

const itineraryRepoService = "itinerary"

type PostgresRepository struct {
	db database.Client
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListPublic(ctx context.Context, filter itinerary.ListFilter) (itinerary.ListPage, error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return itinerary.ListPage{}, itinerary.ErrDependencyUnavailable
	}

	where, args := buildListWhere(filter)

	var total int
	if err := r.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*)
		FROM itineraries i
		WHERE `+where, args...).Scan(&total); err != nil {
		return itinerary.ListPage{}, mapDBError(ctx, "itineraries.list.count", err)
	}

	queryArgs := append([]any(nil), args...)
	limitParam := len(queryArgs) + 1
	queryArgs = append(queryArgs, filter.Limit)
	offsetParam := len(queryArgs) + 1
	queryArgs = append(queryArgs, filter.Offset)

	rows, err := r.db.Pool().Query(ctx, fmt.Sprintf(`
		SELECT
			i.id::text,
			i.title,
			i.slug,
			i.province,
			i.duration_days,
			i.budget_min,
			i.budget_max,
			i.travel_style,
			i.transportation,
			i.cover_image_url,
			COUNT(s.id)::int AS stop_count
		FROM itineraries i
		LEFT JOIN itinerary_stops s ON s.itinerary_id = i.id
		WHERE %s
		GROUP BY
			i.id,
			i.title,
			i.slug,
			i.province,
			i.duration_days,
			i.budget_min,
			i.budget_max,
			i.travel_style,
			i.transportation,
			i.cover_image_url,
			i.created_at
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitParam, offsetParam), queryArgs...)
	if err != nil {
		return itinerary.ListPage{}, mapDBError(ctx, "itineraries.list", err)
	}
	defer rows.Close()

	items := make([]itinerary.ListItem, 0)
	for rows.Next() {
		item, err := scanListItem(rows)
		if err != nil {
			return itinerary.ListPage{}, mapDBError(ctx, "itineraries.list.scan", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return itinerary.ListPage{}, mapDBError(ctx, "itineraries.list.iterate", err)
	}

	return itinerary.ListPage{
		Items: items,
		Page:  filter.Page,
		Limit: filter.Limit,
		Total: total,
	}, nil
}

func (r *PostgresRepository) GetPublicBySlug(ctx context.Context, slug string) (itinerary.Detail, error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return itinerary.Detail{}, itinerary.ErrDependencyUnavailable
	}

	detail, err := scanDetail(r.db.Pool().QueryRow(ctx, `
		SELECT
			i.id::text,
			i.title,
			i.slug,
			i.province,
			i.duration_days,
			i.budget_min,
			i.budget_max,
			i.travel_style,
			i.transportation,
			i.start_location,
			i.description,
			i.cover_image_url
		FROM itineraries i
		WHERE i.is_public = TRUE
			AND i.slug = $1
		LIMIT 1
	`, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return itinerary.Detail{}, itinerary.ErrNotFound
		}
		return itinerary.Detail{}, mapDBError(ctx, "itineraries.get_by_slug", err)
	}

	stops, err := r.loadStops(ctx, detail.ID)
	if err != nil {
		return itinerary.Detail{}, err
	}
	detail.Days = groupStopsByDay(stops)

	return detail, nil
}

func (r *PostgresRepository) loadStops(ctx context.Context, itineraryID string) ([]itinerary.Stop, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT
			s.id::text,
			s.location_id::text,
			l.name,
			l.latitude,
			l.longitude,
			COALESCE(to_char(s.start_time, 'HH24:MI'), '') AS start_time,
			COALESCE(to_char(s.end_time, 'HH24:MI'), '') AS end_time,
			s.note,
			s.estimated_cost,
			s.stop_order,
			s.day_number
		FROM itinerary_stops s
		INNER JOIN locations l ON l.id = s.location_id
		WHERE s.itinerary_id::text = $1
		ORDER BY s.day_number ASC, s.stop_order ASC
	`, itineraryID)
	if err != nil {
		return nil, mapDBError(ctx, "itineraries.stops.list", err)
	}
	defer rows.Close()

	stops := make([]itinerary.Stop, 0)
	for rows.Next() {
		stop, err := scanStop(rows)
		if err != nil {
			return nil, mapDBError(ctx, "itineraries.stops.scan", err)
		}
		stops = append(stops, stop)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, "itineraries.stops.iterate", err)
	}

	return stops, nil
}

func buildListWhere(filter itinerary.ListFilter) (string, []any) {
	conditions := []string{"i.is_public = TRUE"}
	args := make([]any, 0, 4)

	if filter.Province != "" {
		args = append(args, filter.Province)
		conditions = append(conditions, fmt.Sprintf("i.province ILIKE $%d", len(args)))
	}
	if filter.DurationDays > 0 {
		args = append(args, filter.DurationDays)
		conditions = append(conditions, fmt.Sprintf("i.duration_days = $%d", len(args)))
	}
	if filter.BudgetMax > 0 {
		args = append(args, filter.BudgetMax)
		conditions = append(conditions, fmt.Sprintf("i.budget_max <= $%d", len(args)))
	}
	if filter.TravelStyle != "" {
		args = append(args, "%"+filter.TravelStyle+"%")
		conditions = append(conditions, fmt.Sprintf("i.travel_style ILIKE $%d", len(args)))
	}

	return strings.Join(conditions, " AND "), args
}

func groupStopsByDay(stops []itinerary.Stop) []itinerary.Day {
	days := make([]itinerary.Day, 0)
	for _, stop := range stops {
		if len(days) == 0 || days[len(days)-1].DayNumber != stop.DayNumber {
			days = append(days, itinerary.Day{
				DayNumber: stop.DayNumber,
				Stops:     []itinerary.Stop{},
			})
		}
		days[len(days)-1].Stops = append(days[len(days)-1].Stops, stop)
	}
	return days
}

func scanListItem(scanner orm.Scanner) (itinerary.ListItem, error) {
	var item itinerary.ListItem
	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&item.Slug,
		&item.Province,
		&item.DurationDays,
		&item.BudgetMin,
		&item.BudgetMax,
		&item.TravelStyle,
		&item.Transportation,
		&item.CoverImageURL,
		&item.StopCount,
	)
	return item, err
}

func scanDetail(scanner orm.Scanner) (itinerary.Detail, error) {
	var item itinerary.Detail
	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&item.Slug,
		&item.Province,
		&item.DurationDays,
		&item.BudgetMin,
		&item.BudgetMax,
		&item.TravelStyle,
		&item.Transportation,
		&item.StartLocation,
		&item.Description,
		&item.CoverImageURL,
	)
	if item.Days == nil {
		item.Days = []itinerary.Day{}
	}
	return item, err
}

func scanStop(scanner orm.Scanner) (itinerary.Stop, error) {
	var item itinerary.Stop
	err := scanner.Scan(
		&item.ID,
		&item.LocationID,
		&item.LocationName,
		&item.Latitude,
		&item.Longitude,
		&item.StartTime,
		&item.EndTime,
		&item.Note,
		&item.EstimatedCost,
		&item.StopOrder,
		&item.DayNumber,
	)
	return item, err
}

func mapDBError(ctx context.Context, operation string, err error) error {
	return share.MapDBError(ctx, itineraryRepoService, operation, err, itinerary.ErrDependencyUnavailable, itinerary.ErrInternal)
}

var _ itinerary.Repository = (*PostgresRepository)(nil)
