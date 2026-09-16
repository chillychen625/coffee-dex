// Command migrate copies all data from the local SQLite database into the
// Neon Postgres database, in one transaction. Re-running is safe: inserts
// use ON CONFLICT DO NOTHING.
//
// Run from the repo root: go run ./scripts
//
// Use -wipe to first delete all existing rows from the Postgres tables
// (e.g. after migrating stale data by mistake, or to start over).
// go run ./scripts -wipe
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver

	"github.com/jackc/pgx/v5/pgxpool"

	"go-coffee-log/storage"
)

// FK-safe order: parents before children.
var tables = []string{"coffees", "pokemons", "brews", "brewers", "coffee_pokemon"}

func main() {
	wipe := flag.Bool("wipe", false, "delete all existing rows from the Postgres tables before migrating")
	flag.Parse()

	ctx := context.Background()

	pgDB, err := storage.NewPGDB(ctx) // connects and creates tables
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open Postgres: %v\n", err)
		os.Exit(1)
	}
	defer pgDB.Close()

	if *wipe {
		// All referencing tables must be listed in the same TRUNCATE —
		// truncating "coffees" alone would fail, since brews and
		// coffee_pokemon hold foreign keys into it.
		if _, err := pgDB.Pool().Exec(ctx, "TRUNCATE TABLE "+strings.Join(tables, ", ")); err != nil {
			fmt.Fprintf(os.Stderr, "failed to wipe tables: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("wiped all tables")
	}

	// Path is relative to where `go run` executes (the repo root).
	sqliteDB, err := sql.Open("sqlite", "coffee-dex.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open SQLite database: %v\n", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()
	if err := sqliteDB.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping SQLite database: %v\n", err)
		os.Exit(1)
	}

	// ONE transaction for the whole migration: a failure anywhere leaves
	// Neon empty rather than half-full. Begin checks out one pooled
	// connection, sends BEGIN, and hands us a handle pinned to it.
	tx, err := pgDB.Pool().Begin(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to begin transaction: %v\n", err)
		os.Exit(1)
	}
	// Rollback after a Commit is a no-op error, so this is a pure safety net.
	defer tx.Rollback(ctx)

	for _, t := range tables {
		if err := migrateTable(ctx, sqliteDB, tx, t); err != nil {
			fmt.Fprintf(os.Stderr, "migrating %s: %v\n", t, err)
			os.Exit(1) // deferred Rollback unwinds everything
		}
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to commit: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nVerifying row counts (sqlite vs postgres):")
	for _, t := range tables {
		srcCount, err := countSQLite(sqliteDB, t)
		if err != nil {
			fmt.Fprintf(os.Stderr, "count %s in sqlite: %v\n", t, err)
			os.Exit(1)
		}
		dstCount, err := countPostgres(ctx, pgDB.Pool(), t)
		if err != nil {
			fmt.Fprintf(os.Stderr, "count %s in postgres: %v\n", t, err)
			os.Exit(1)
		}
		match := "OK"
		if srcCount != dstCount {
			match = "MISMATCH"
		}
		fmt.Printf("  %-16s %d vs %d  %s\n", t, srcCount, dstCount, match)
	}

	fmt.Println("\ndone")
}

// migrateTable copies every row of one table from sqlite into the open
// Postgres transaction. It is table-agnostic: it discovers the columns
// from the query result and builds the INSERT dynamically.
func migrateTable(ctx context.Context, src *sql.DB, dst pgx.Tx, table string) error {
	// Table names cannot be passed as placeholders (placeholders are for
	// VALUES only) — but `table` comes from our own hardcoded list, so
	// plain concatenation is safe here.
	rows, err := src.Query("SELECT * FROM " + table)
	if err != nil {
		return fmt.Errorf("select: %w", err)
	}
	defer rows.Close()

	// Discover the columns and build a scan destination per column.
	//
	// ScanType() tells us the Go type the driver produces (string for
	// TEXT, int64 for INTEGER, float64 for REAL). We allocate **T, not
	// *T: database/sql writes a non-nil *T for values and a nil *T for
	// SQL NULL — and pgx encodes a nil pointer as NULL, so nulls survive
	// the trip instead of degrading to "" or 0.
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("column types: %w", err)
	}

	colNames := make([]string, len(colTypes))
	dests := make([]any, len(colTypes))
	for i, ct := range colTypes {
		colNames[i] = ct.Name()
		// The modernc sqlite driver returns a nil ScanType for columns it
		// can't statically type (sqlite is dynamically typed). Fall back
		// to plain `any`: database/sql puts the raw driver value in it
		// (string / int64 / float64, or nil for NULL) — all of which pgx
		// can encode as-is.
		st := ct.ScanType()
		if st == nil {
			dests[i] = new(any)
			continue
		}
		dests[i] = reflect.New(reflect.PointerTo(st)).Interface()
	}

	// INSERT INTO coffees (id, name, ...) VALUES ($1, $2, ...) ON CONFLICT DO NOTHING
	placeholders := make([]string, len(colTypes))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	insertSQL := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
		table, strings.Join(colNames, ", "), strings.Join(placeholders, ", "),
	)

	var inserted int64
	for rows.Next() {
		if err := rows.Scan(dests...); err != nil {
			return fmt.Errorf("scan: %w", err)
		}

		// dests[i] is a **T; strip one level so the arg is a *T (or nil
		// for NULL), which pgx handles natively.
		args := make([]any, len(dests))
		for i, d := range dests {
			args[i] = reflect.ValueOf(d).Elem().Interface()
		}

		ct, err := dst.Exec(ctx, insertSQL, args...)
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		inserted += ct.RowsAffected()
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate: %w", err)
	}

	fmt.Printf("  %-16s %d rows inserted\n", table, inserted)
	return nil
}

// countSQLite returns the number of rows in a table in the sqlite database.
// database/sql's QueryRow has no ctx parameter (a pre-contexts API); the
// context-aware variant is a separate method, QueryRowContext.
func countSQLite(db *sql.DB, table string) (int64, error) {
	var n int64
	// Same rule as everywhere: identifiers are concatenated (from our own
	// list), COUNT(*) takes no values, so no placeholder at all.
	err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n)
	return n, err
}

// countPostgres returns the number of rows in a table in Postgres.
func countPostgres(ctx context.Context, pool *pgxpool.Pool, table string) (int64, error) {
	var n int64
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n)
	return n, err
}
