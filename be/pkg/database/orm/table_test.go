package orm

import "testing"

func TestBuildInsertSortsColumnsAndReturnsArgs(t *testing.T) {
	sql, args := buildInsert("categories", Values{
		"slug": "beach",
		"name": "Beach",
	}, InsertOptions{})

	expectedSQL := `INSERT INTO "categories" ("name", "slug") VALUES ($1, $2)`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
	if len(args) != 2 || args[0] != "Beach" || args[1] != "beach" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildInsertReturning(t *testing.T) {
	sql, _ := buildInsert("images", Values{"url": "https://example.com"}, InsertOptions{Returning: []string{"id", "created_at"}})

	expectedSQL := `INSERT INTO "images" ("url") VALUES ($1) RETURNING "id", "created_at"`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
}

func TestBuildInsertOnConflictDoNothing(t *testing.T) {
	sql, _ := buildInsert("notifications", Values{"id": "1"}, InsertOptions{OnConflictDoNothing: true})

	expectedSQL := `INSERT INTO "notifications" ("id") VALUES ($1) ON CONFLICT DO NOTHING`
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

func TestBuildUpdateWhere(t *testing.T) {
	sql, args := buildUpdateWhere("users", Values{
		"updated_at":    "now",
		"password_hash": "hash",
	}, "id = $3 AND is_active = TRUE", 0, uint64(9))

	expectedSQL := `UPDATE "users" SET "password_hash" = $1, "updated_at" = $2 WHERE id = $3 AND is_active = TRUE`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
	if len(args) != 3 || args[0] != "hash" || args[1] != "now" || args[2] != uint64(9) {
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

func TestSelectSQLKeepsStaticExpressionsRaw(t *testing.T) {
	table := NewTable[string](nil, "posts p", []string{"p.id::text", "COALESCE(p.caption, '')"}, nil)

	sql := table.selectSQL(QueryOptions{Where: "p.location_id::text = $1"})

	expectedSQL := `SELECT p.id::text, COALESCE(p.caption, '') FROM posts p WHERE p.location_id::text = $1`
	if sql != expectedSQL {
		t.Fatalf("expected sql %q, got %q", expectedSQL, sql)
	}
}
