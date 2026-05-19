package orm

import "testing"

func TestBuildInsertSortsColumnsAndReturnsArgs(t *testing.T) {
	sql, args := buildInsert("categories", Values{
		"slug": "beach",
		"name": "Beach",
	}, nil)

	expectedSQL := `INSERT INTO "categories" ("name", "slug") VALUES ($1, $2)`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
	if len(args) != 2 || args[0] != "Beach" || args[1] != "beach" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildInsertReturning(t *testing.T) {
	sql, _ := buildInsert("images", Values{"url": "https://example.com"}, []string{"id", "created_at"})

	expectedSQL := `INSERT INTO "images" ("url") VALUES ($1) RETURNING "id", "created_at"`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
}

func TestBuildUpdateByID(t *testing.T) {
	sql, args := buildUpdateByID("categories", Values{
		"slug": "city",
		"name": "City",
	}, uint64(7))

	expectedSQL := `UPDATE "categories" SET "name" = $1, "slug" = $2 WHERE "id" = $3`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
	if len(args) != 3 || args[0] != "City" || args[1] != "city" || args[2] != uint64(7) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestSelectSQL(t *testing.T) {
	table := NewTable[string](nil, "categories", []string{"id", "name", "slug"}, nil)

	sql := table.selectSQL(QueryOptions{
		Where:   "slug = $1",
		OrderBy: `"name" ASC`,
		Limit:   10,
	})

	expectedSQL := `SELECT "id", "name", "slug" FROM "categories" WHERE slug = $1 ORDER BY "name" ASC LIMIT 10`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
}
