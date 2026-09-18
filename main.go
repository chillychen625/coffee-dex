package main

import (
	"context"
	"go-coffee-log/game"
	"go-coffee-log/service"
	"go-coffee-log/storage"
	"log"
)

func main() {
	ctx := context.Background()

	// Connect to Postgres (Neon). Reads DATABASE_URL from the environment
	// or a .env file next to the binary, and creates tables if missing.
	pgDB, err := storage.NewPGDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgDB.Close()

	// Storage layer
	coffeeStore := storage.NewPgCoffeeStorage(pgDB.Pool())
	brewStore := storage.NewPgBrewStorage(pgDB.Pool())
	brewerStore := storage.NewPgBrewerStorage(pgDB.Pool())
	pokemonStore := storage.NewPgPokemonStorage(pgDB.Pool())

	// Service layer
	coffeeService := service.NewCoffeeService(ctx, coffeeStore)
	brewService := service.NewBrewService(ctx, brewStore, coffeeStore)

	pokemonService := service.NewPokemonService(ctx, pokemonStore, coffeeService, brewService)
	if err := pokemonService.InitializePokemonData(); err != nil {
		log.Printf("Warning: failed to initialize Pokemon data: %v", err)
	}

	brewerService := service.NewBrewerService(ctx, brewerStore)
	statisticsService := service.NewStatisticsService(ctx, coffeeStore, brewStore, pokemonStore)

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
