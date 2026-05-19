package orm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Queryer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Scanner interface {
	Scan(dest ...any) error
}

type Mapper[T any] func(Scanner) (T, error)

type Values map[string]any

type QueryOptions struct {
	Where   string
	Args    []any
	OrderBy string
	Limit   int
}

type Table[T any] struct {
	db      Queryer
	name    string
	columns []string
	mapper  Mapper[T]
}

func NewTable[T any](db Queryer, name string, columns []string, mapper Mapper[T]) *Table[T] {
	return &Table[T]{
		db:      db,
		name:    name,
		columns: append([]string(nil), columns...),
		mapper:  mapper,
	}
}

func (t *Table[T]) Insert(ctx context.Context, values Values) (pgconn.CommandTag, error) {
	sql, args := buildInsert(t.name, values, nil)
	return t.db.Exec(ctx, sql, args...)
}

func (t *Table[T]) InsertReturning(ctx context.Context, values Values, returning []string, destinations ...any) error {
	sql, args := buildInsert(t.name, values, returning)
	return t.db.QueryRow(ctx, sql, args...).Scan(destinations...)
}

func (t *Table[T]) FindByID(ctx context.Context, id any) (T, error) {
	return t.FindOne(ctx, "id = $1", id)
}

func (t *Table[T]) FindOne(ctx context.Context, where string, args ...any) (T, error) {
	var zero T
	sql := t.selectSQL(QueryOptions{Where: where})
	item, err := t.mapper(t.db.QueryRow(ctx, sql, args...))
	if err != nil {
		return zero, err
	}

	return item, nil
}

func (t *Table[T]) List(ctx context.Context, options QueryOptions) ([]T, error) {
	sql := t.selectSQL(options)
	rows, err := t.db.Query(ctx, sql, options.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]T, 0)
	for rows.Next() {
		item, err := t.mapper(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (t *Table[T]) UpdateByID(ctx context.Context, id any, values Values) (pgconn.CommandTag, error) {
	sql, args := buildUpdateByID(t.name, values, id)
	return t.db.Exec(ctx, sql, args...)
}

func (t *Table[T]) DeleteByID(ctx context.Context, id any) (pgconn.CommandTag, error) {
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", quoteIdent(t.name), quoteIdent("id"))
	return t.db.Exec(ctx, sql, id)
}

func (t *Table[T]) selectSQL(options QueryOptions) string {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(quoteIdentList(t.columns))
	builder.WriteString(" FROM ")
	builder.WriteString(quoteIdent(t.name))

	if strings.TrimSpace(options.Where) != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(options.Where)
	}
	if strings.TrimSpace(options.OrderBy) != "" {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(options.OrderBy)
	}
	if options.Limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", options.Limit))
	}

	return builder.String()
}

func buildInsert(table string, values Values, returning []string) (string, []any) {
	keys := sortedKeys(values)
	columns := make([]string, 0, len(keys))
	placeholders := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		columns = append(columns, quoteIdent(key))
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		args = append(args, values[key])
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(table),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	if len(returning) > 0 {
		sql += " RETURNING " + quoteIdentList(returning)
	}

	return sql, args
}

func buildUpdateByID(table string, values Values, id any) (string, []any) {
	keys := sortedKeys(values)
	assignments := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys)+1)
	for index, key := range keys {
		assignments = append(assignments, fmt.Sprintf("%s = $%d", quoteIdent(key), index+1))
		args = append(args, values[key])
	}
	args = append(args, id)

	sql := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		quoteIdent(table),
		strings.Join(assignments, ", "),
		quoteIdent("id"),
		len(args),
	)

	return sql, args
}

func sortedKeys(values Values) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func quoteIdentList(columns []string) string {
	quoted := make([]string, 0, len(columns))
	for _, column := range columns {
		quoted = append(quoted, quoteIdent(column))
	}
	return strings.Join(quoted, ", ")
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
