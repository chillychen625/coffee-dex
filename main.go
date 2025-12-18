package main

import (
	"flag"
	"fmt"
	"go-coffee-log/handlers"
	"go-coffee-log/service"
	"go-coffee-log/storage"
	"log"
	"net/http"
	"os"
)

func main() {
	// Command-line flags for storage configuration
	storageType := flag.String("storage", "memory", "Storage type: memory or mysql")
	mysqlHost := flag.String("mysql-host", "localhost:3306", "MySQL host")
	mysqlUser := flag.String("mysql-user", "root", "MySQL user")
	mysqlPassword := flag.String("mysql-password", "", "MySQL password")
	mysqlDB := flag.String("mysql-db", "coffee_log", "MySQL database name")
	flag.Parse()

	// Initialize storage based on flag
	var store storage.CoffeeStorage
	var err error

	switch *storageType {
	case "mysql":
		store, err = storage.NewMySQLStorage(*mysqlHost, *mysqlUser, *mysqlPassword, *mysqlDB)
		if err != nil {
			log.Fatalf("Failed to initialize MySQL storage: %v", err)
		}
		fmt.Println("Using MySQL storage")
		
		// Close MySQL connection on shutdown
		if mysqlStore, ok := store.(*storage.MySQLStorage); ok {
			defer mysqlStore.Close()
		}
	case "memory":
		store = storage.NewMemoryStorage()
		fmt.Println("Using in-memory storage")
	default:
		fmt.Fprintf(os.Stderr, "Invalid storage type: %s. Use 'memory' or 'mysql'\n", *storageType)
		os.Exit(1)
	}

	coffeeService := service.NewCoffeeService(store)
	coffeeHandler := handlers.NewCoffeeHandler(coffeeService)
	mux := http.NewServeMux()

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
	
	// Route to /coffees/{id}
	mux.HandleFunc("/coffees/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	
	// Add catch-all handler LAST
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	
	loggedMux := loggingMiddleware(mux)
	
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s", r.Method, r.URL.Path)
	})
}