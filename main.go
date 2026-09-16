package main

import (
	"flag"
	"go-coffee-log/game"
	"go-coffee-log/service"
	"go-coffee-log/storage"
	"log"
)

func main() {

	dbPath := flag.String("db", "./coffee-dex.db", "SQLite database file path")
	flag.Parse()

	// Open SQLite database (creates file if it doesn't exist)
	sqliteDB, err := storage.NewSQLiteDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database %s: %v", *dbPath, err)
	}
	defer sqliteDB.Close()
	db := sqliteDB.DB()

	// Storage layer
	coffeeStore := storage.NewSQLiteCoffeeStorage(db)
	brewStore := storage.NewSQLiteBrewStorage(db)
	brewerStore := storage.NewSQLiteBrewerStorage(db)
	pokemonStore := storage.NewSQLitePokemonStorage(db)

	// Service layer
	coffeeService := service.NewCoffeeService(coffeeStore)
	brewService := service.NewBrewService(brewStore, coffeeStore)

	pokemonService := service.NewPokemonService(pokemonStore, coffeeService, brewService)
	if err := pokemonService.InitializePokemonData(); err != nil {
		log.Printf("Warning: failed to initialize Pokemon data: %v", err)
	}

	brewerService := service.NewBrewerService(brewerStore)
	statisticsService := service.NewStatisticsService(coffeeStore, brewStore, pokemonStore)

	// Launch game
	svc := &game.Services{
		Coffee:     coffeeService,
		Brew:       brewService,
		Brewer:     brewerService,
		Pokemon:    pokemonService,
		Statistics: statisticsService,
	}

	if err := game.Run(svc); err != nil {
		log.Fatalf("Game error: %v", err)
	}
}
