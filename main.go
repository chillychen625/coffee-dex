package main

import (
	"database/sql"
	"flag"
	"fmt"
	"go-coffee-log/handlers"
	"go-coffee-log/service"
	"go-coffee-log/storage"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Command-line flags for storage configuration
	storageType := flag.String("storage", "memory", "Storage type: memory or mysql")
	mysqlHost := flag.String("mysql-host", "localhost:3306", "MySQL host")
	mysqlUser := flag.String("mysql-user", "root", "MySQL user")
	mysqlPassword := flag.String("mysql-password", "", "MySQL password")
	mysqlDB := flag.String("mysql-db", "coffee_log", "MySQL database name")
	
	// Pokemon configuration flags
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama base URL")
	ollamaModel := flag.String("ollama-model", "qwen3:4b", "Ollama model name")
	enableLLM := flag.Bool("enable-llm", true, "Enable LLM Pokemon mapping")
	
	flag.Parse()

	// Initialize storage based on flag
	var store storage.CoffeeStorage
	var pokemonStorage storage.PokemonStorage
	var db *sql.DB
	var err error

	switch *storageType {
	case "mysql":
		store, err = storage.NewMySQLStorage(*mysqlHost, *mysqlUser, *mysqlPassword, *mysqlDB)
		if err != nil {
			log.Fatalf("Failed to initialize MySQL storage: %v", err)
		}
		fmt.Println("Using MySQL storage")
		
		// Get the underlying database connection for Pokemon storage
		if mysqlStore, ok := store.(*storage.MySQLStorage); ok {
			// Access the private db field - we'll need to modify MySQLStorage to expose this
			// For now, we'll create a new connection
			log.Printf("INFO: Opening MySQL connection for Pokemon/Brewer storage")
			db, err = openMySQLConnection(*mysqlHost, *mysqlUser, *mysqlPassword, *mysqlDB)
			if err != nil {
				log.Fatalf("Failed to create Pokemon DB connection: %v", err)
			}
			
			// Test the connection
			if err := db.Ping(); err != nil {
				log.Fatalf("Failed to ping Pokemon DB connection: %v", err)
			}
			log.Printf("INFO: MySQL connection for Pokemon/Brewer storage successful")
			
			pokemonStorage = storage.NewMySQLPokemonStorage(db)
			
			defer mysqlStore.Close()
			defer db.Close()
		}
	case "memory":
		store = storage.NewMemoryStorage()
		fmt.Println("Using in-memory storage")
		// Pokemon storage not available with memory storage
		pokemonStorage = nil
	default:
		fmt.Fprintf(os.Stderr, "Invalid storage type: %s. Use 'memory' or 'mysql'\n", *storageType)
		os.Exit(1)
	}

	// Initialize services
	coffeeService := service.NewCoffeeService(store)
	
	// Initialize statistics service
	var statisticsService *service.StatisticsService
	
	// Initialize brewer service
	var brewerService *service.BrewerService
	var brewerStorage storage.BrewerStorage
	
	// Initialize Pokemon service
	var pokemonService *service.PokemonService
	var llmService *service.LLMService
	
	if pokemonStorage != nil {
		if *enableLLM {
			llmService = service.NewLLMService(*ollamaURL, *ollamaModel)
			// Test LLM connection
			if err := llmService.TestConnection(); err != nil {
				log.Printf("Warning: LLM service connection failed: %v", err)
				llmService = nil
			} else {
				fmt.Println("LLM service connected successfully")
			}
		}
		
		pokemonService = service.NewPokemonService(pokemonStorage, coffeeService, llmService)
		
		// Initialize Pokemon data
		if err := pokemonService.InitializePokemonData(); err != nil {
			log.Printf("Failed to initialize Pokemon data: %v", err)
		}
		
		// Initialize statistics service (requires Pokemon storage)
		statisticsService = service.NewStatisticsService(store, pokemonStorage)
		
		// Initialize brewer service (requires MySQL storage)
		log.Printf("INFO: Initializing brewer storage with MySQL connection")
		brewerStorage = storage.NewMySQLBrewerStorage(db, store)
		brewerService = service.NewBrewerService(brewerStorage)
		log.Printf("INFO: Brewer service initialized successfully")
	} else {
		fmt.Println("Pokemon features disabled (requires MySQL storage)")
	}
	
	// Initialize handlers
	coffeeHandler := handlers.NewCoffeeHandler(coffeeService)
	
	var pokemonHandler *handlers.PokemonHandler
	var statisticsHandler *handlers.StatisticsHandler
	var brewerHandler *handlers.BrewerHandler
	
	if pokemonService != nil {
		pokemonHandler = handlers.NewPokemonHandler(pokemonService, coffeeService)
	}
	
	if statisticsService != nil {
		statisticsHandler = handlers.NewStatisticsHandler(statisticsService)
	}
	
	if brewerService != nil {
		brewerHandler = handlers.NewBrewerHandler(brewerService)
	}
	
	mux := http.NewServeMux()

	// Coffee routes
	mux.HandleFunc("/coffees/recent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			coffeeHandler.GetRecentCoffees(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	mux.HandleFunc("/coffees", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			coffeeHandler.CreateCoffee(w, r)
		case http.MethodGet:
			coffeeHandler.ListCoffees(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	// Pokemon routes (if Pokemon service is available)
	if pokemonHandler != nil {
		// Pokemon routes for a specific coffee
		mux.HandleFunc("/pokemon/", func(w http.ResponseWriter, r *http.Request) {
			// Extract coffee_id from path: /pokemon/{coffee_id}
			path := strings.TrimPrefix(r.URL.Path, "/pokemon/")
			parts := strings.Split(path, "/")
			if len(parts) == 0 || parts[0] == "" {
				http.NotFound(w, r)
				return
			}
			
			coffeeID := parts[0]
			
			// Handle /pokemon/{coffee_id}/nickname
			if len(parts) == 2 && parts[1] == "nickname" {
				if r.Method == http.MethodPut {
					// Temporarily set PathValue for handler
					r.SetPathValue("coffee_id", coffeeID)
					pokemonHandler.UpdateNickname(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			// Handle /pokemon/{coffee_id}
			if len(parts) == 1 {
				r.SetPathValue("coffee_id", coffeeID)
				switch r.Method {
				case http.MethodPost:
					pokemonHandler.GeneratePokemon(w, r)
				case http.MethodGet:
					pokemonHandler.GetCoffeePokemon(w, r)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}
			
			http.NotFound(w, r)
		})
		
		// CoffeeDex routes
		mux.HandleFunc("/pokedex/stats", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				pokemonHandler.GetPokemonStats(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
		
		mux.HandleFunc("/pokedex", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				pokemonHandler.GetCoffeeDex(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
	}
	
	// Statistics routes (if statistics service is available)
	if statisticsHandler != nil {
		mux.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				statisticsHandler.GetStatistics(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
	}
	
	// Brewer routes (if brewer service is available)
	if brewerHandler != nil {
		mux.HandleFunc("/brewers/pokeball-types", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				brewerHandler.GetAvailablePokeballTypes(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
		
		
		mux.HandleFunc("/brewers", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				brewerHandler.CreateBrewer(w, r)
			case http.MethodGet:
				brewerHandler.GetAllBrewers(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
		
		mux.HandleFunc("/brewers/", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/brewers/")
			parts := strings.Split(path, "/")
			if len(parts) == 0 || parts[0] == "" {
				http.NotFound(w, r)
				return
			}
			
			brewerID := parts[0]
			
			
			// Handle /brewers/{id}/standalone-recipes/{recipe_id}
			if len(parts) == 3 && parts[1] == "standalone-recipes" {
				r.SetPathValue("id", brewerID)
				r.SetPathValue("recipe_id", parts[2])
				if r.Method == http.MethodDelete {
					brewerHandler.RemoveStandaloneRecipe(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			// Handle /brewers/{id}/standalone-recipes
			if len(parts) == 2 && parts[1] == "standalone-recipes" {
				r.SetPathValue("id", brewerID)
				if r.Method == http.MethodPost {
					brewerHandler.AddStandaloneRecipe(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			
			// Handle /brewers/{id}
			if len(parts) == 1 {
				r.SetPathValue("id", brewerID)
				if r.Method == http.MethodDelete {
					brewerHandler.DeleteBrewer(w, r)
					return
				}
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			http.NotFound(w, r)
		})
	}
	
	// Route to /coffees/{id}
	mux.HandleFunc("/coffees/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/coffees/")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}
		
		r.SetPathValue("id", id)
		switch r.Method {
		case http.MethodGet:
			coffeeHandler.GetCoffee(w, r)
		case http.MethodPut:
			coffeeHandler.UpdateCoffee(w, r)
		case http.MethodDelete:
			coffeeHandler.DeleteCoffee(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	
	// Static file server for Pokemon sprites
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	
	// Add catch-all handler LAST
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	
	loggedMux := loggingMiddleware(mux)
	
	fmt.Println("Server starting on :8080")
	if pokemonService != nil {
		fmt.Println("Pokemon features enabled")
	} else {
		fmt.Println("Pokemon features disabled")
	}
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}

// openMySQLConnection opens a MySQL database connection
func openMySQLConnection(host, user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, dbname)
	return sql.Open("mysql", dsn)
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s", r.Method, r.URL.Path)
	})
}