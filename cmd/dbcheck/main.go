// Command dbcheck is a throwaway smoke test for the Neon migration:
// it connects using storage.NewPGDB (which also creates the tables)
// and lists what exists.
package main

import (
	"context"
	"fmt"
	"os"

	"go-coffee-log/storage"
)

func main() {
	ctx := context.Background()

	pgDB, err := storage.NewPGDB(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	defer pgDB.Close()

	rows, err := pgDB.Pool().Query(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: listing tables: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Connected to Neon. Tables in public schema:")
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: scan: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  -", name)
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
}
