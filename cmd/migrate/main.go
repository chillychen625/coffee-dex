// Migration tool: copies data from MySQL to a new SQLite database.
// Usage: go run ./cmd/migrate -mysql-password=<pw> [-sqlite=./coffee-dex.db]
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go-coffee-log/storage"
)

func main() {
	mysqlHost := flag.String("mysql-host", "localhost:3306", "MySQL host")
	mysqlUser := flag.String("mysql-user", "root", "MySQL user")
	mysqlPassword := flag.String("mysql-password", "", "MySQL password")
	mysqlDB := flag.String("mysql-db", "coffee_log", "MySQL database name")
	sqlitePath := flag.String("sqlite", "./coffee-dex.db", "SQLite output path")
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", *mysqlUser, *mysqlPassword, *mysqlHost, *mysqlDB)
	mysql, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer mysql.Close()
	if err := mysql.Ping(); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}

	lite, err := storage.NewSQLiteDB(*sqlitePath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer lite.Close()
	db := lite.DB()

	migrateCoffees(mysql, db)
	migrateBrews(mysql, db)
	migrateBrewers(mysql, db)
	migratePokemons(mysql, db)
	migrateCoffeePokemon(mysql, db)

	fmt.Println("Migration complete.")
}

func migrateCoffees(src, dst *sql.DB) {
	rows, err := src.Query(`SELECT id, name, origin, roaster, variety, roast_level, processing_method,
		roast_date, is_finished, created_at, updated_at FROM coffees`)
	if err != nil {
		log.Fatalf("query coffees: %v", err)
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, name string
		var origin, roaster, variety, roastLevel, processingMethod sql.NullString
		var roastDate sql.NullTime
		var isFinished sql.NullBool
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &origin, &roaster, &variety, &roastLevel, &processingMethod,
			&roastDate, &isFinished, &createdAt, &updatedAt); err != nil {
			log.Printf("scan coffee: %v", err)
			continue
		}

		var roastDateStr interface{}
		if roastDate.Valid {
			roastDateStr = roastDate.Time.Format("2006-01-02")
		}
		finished := 0
		if isFinished.Valid && isFinished.Bool {
			finished = 1
		}

		_, err := dst.Exec(`INSERT OR IGNORE INTO coffees
			(id, name, origin, roaster, variety, roast_level, processing_method, roast_date, is_finished, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			id, name, origin, roaster, variety, roastLevel, processingMethod,
			roastDateStr, finished,
			createdAt.UTC().Format(time.RFC3339),
			updatedAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			log.Printf("insert coffee %s: %v", id, err)
			continue
		}
		n++
	}
	fmt.Printf("  coffees: %d rows\n", n)
}

func migrateBrews(src, dst *sql.DB) {
	rows, err := src.Query(`SELECT id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper,
		end_time_minutes, end_time_seconds, created_at FROM brews`)
	if err != nil {
		log.Fatalf("query brews: %v", err)
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, coffeeID string
		var tastingNotes, tastingTraits, recipe []byte
		var rating, endMins, endSecs sql.NullInt64
		var dripper sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &coffeeID, &tastingNotes, &tastingTraits, &rating, &recipe,
			&dripper, &endMins, &endSecs, &createdAt); err != nil {
			log.Printf("scan brew: %v", err)
			continue
		}

		// Normalize JSON — store as compact string
		tastingNotes = normalizeJSON(tastingNotes)
		tastingTraits = normalizeJSON(tastingTraits)
		recipe = normalizeJSON(recipe)

		_, err := dst.Exec(`INSERT OR IGNORE INTO brews
			(id, coffee_id, tasting_notes, tasting_traits, rating, recipe, dripper, end_time_minutes, end_time_seconds, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			id, coffeeID, string(tastingNotes), string(tastingTraits), rating, string(recipe),
			dripper, endMins, endSecs,
			createdAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			log.Printf("insert brew %s: %v", id, err)
			continue
		}
		n++
	}
	fmt.Printf("  brews: %d rows\n", n)
}

func migrateBrewers(src, dst *sql.DB) {
	rows, err := src.Query(`SELECT id, name, pokeball_type, recipes, created_at FROM brewers`)
	if err != nil {
		// Table might not exist on older installs
		log.Printf("query brewers (skipping if missing): %v", err)
		return
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, name, pokeballType string
		var recipes []byte
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &pokeballType, &recipes, &createdAt); err != nil {
			log.Printf("scan brewer: %v", err)
			continue
		}

		_, err := dst.Exec(`INSERT OR IGNORE INTO brewers (id, name, pokeball_type, recipes, created_at)
			VALUES (?,?,?,?,?)`,
			id, name, pokeballType, normalizeJSON(recipes),
			createdAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			log.Printf("insert brewer %s: %v", id, err)
			continue
		}
		n++
	}
	fmt.Printf("  brewers: %d rows\n", n)
}

func migratePokemons(src, dst *sql.DB) {
	rows, err := src.Query(`SELECT id, name, type, sprite_path, base_stats, description FROM pokemons`)
	if err != nil {
		log.Printf("query pokemons (skipping if missing): %v", err)
		return
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id int
		var name, pokemonType, spritePath string
		var baseStats []byte
		var description sql.NullString

		if err := rows.Scan(&id, &name, &pokemonType, &spritePath, &baseStats, &description); err != nil {
			log.Printf("scan pokemon: %v", err)
			continue
		}

		_, err := dst.Exec(`INSERT OR IGNORE INTO pokemons (id, name, type, sprite_path, base_stats, description)
			VALUES (?,?,?,?,?,?)`,
			id, name, pokemonType, spritePath, normalizeJSON(baseStats), description,
		)
		if err != nil {
			log.Printf("insert pokemon %d: %v", id, err)
			continue
		}
		n++
	}
	fmt.Printf("  pokemons: %d rows\n", n)
}

func migrateCoffeePokemon(src, dst *sql.DB) {
	rows, err := src.Query(`SELECT id, coffee_id, pokemon_id, nickname, level,
		mapping_confidence, llm_description, trait_mapping, created_at FROM coffee_pokemon`)
	if err != nil {
		log.Printf("query coffee_pokemon (skipping if missing): %v", err)
		return
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, coffeeID string
		var pokemonID int
		var nickname, llmDesc sql.NullString
		var level sql.NullInt64
		var confidence sql.NullFloat64
		var traitMapping []byte
		var createdAt time.Time

		if err := rows.Scan(&id, &coffeeID, &pokemonID, &nickname, &level,
			&confidence, &llmDesc, &traitMapping, &createdAt); err != nil {
			log.Printf("scan coffee_pokemon: %v", err)
			continue
		}

		_, err := dst.Exec(`INSERT OR IGNORE INTO coffee_pokemon
			(id, coffee_id, pokemon_id, nickname, level, mapping_confidence, llm_description, trait_mapping, created_at)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			id, coffeeID, pokemonID, nickname, level, confidence, llmDesc,
			normalizeJSON(traitMapping),
			createdAt.UTC().Format(time.RFC3339),
		)
		if err != nil {
			log.Printf("insert coffee_pokemon %s: %v", id, err)
			continue
		}
		n++
	}
	fmt.Printf("  coffee_pokemon: %d rows\n", n)
}

func normalizeJSON(b []byte) []byte {
	if len(b) == 0 {
		return []byte("null")
	}
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return b
	}
	out, _ := json.Marshal(v)
	return out
}
