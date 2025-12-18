package main

import (
	"flag"
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"log"
	"time"

	"github.com/google/uuid"
)

// Sample coffee data for testing
var sampleCoffees = []struct {
	name             string
	origin           string
	roaster          string
	roastLevel       string
	processingMethod string
	tastingNotes     [5]string
	tastingTraits    models.TastingTraits
	rating           int
	recipe           []string
	dripper          string
	endTime          models.DrawDownTime
}{
	{
		name:             "Ethiopian Yirgacheffe",
		origin:           "Ethiopia",
		roaster:          "Blue Bottle",
		roastLevel:       "light",
		processingMethod: "washed",
		tastingNotes:     [5]string{"blueberry", "jasmine", "honey", "citrus", "tea"},
		tastingTraits: models.TastingTraits{
			BerryIntensity:        8,
			StonefruitIntensity:   3,
			RoastIntensity:        2,
			CitrusFruitsIntensity: 7,
			Bitterness:            1,
			Florality:             9,
			Spice:                 2,
			Sweetness:             8,
			AromaticIntensity:     9,
			Savory:                1,
			Body:                  5,
			Cleanliness:           9,
		},
		rating:  9,
		recipe:  []string{"20g coffee", "320ml water", "95°C", "V60 pour over"},
		dripper: "Hario V60",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 45},
	},
	{
		name:             "Colombian Supremo",
		origin:           "Colombia",
		roaster:          "Counter Culture",
		roastLevel:       "medium",
		processingMethod: "washed",
		tastingNotes:     [5]string{"chocolate", "caramel", "nuts", "orange", "brown sugar"},
		tastingTraits: models.TastingTraits{
			BerryIntensity:        2,
			StonefruitIntensity:   4,
			RoastIntensity:        5,
			CitrusFruitsIntensity: 5,
			Bitterness:            3,
			Florality:             3,
			Spice:                 2,
			Sweetness:             7,
			AromaticIntensity:     6,
			Savory:                4,
			Body:                  7,
			Cleanliness:           8,
		},
		rating:  8,
		recipe:  []string{"18g coffee", "300ml water", "93°C", "Kalita Wave"},
		dripper: "Kalita Wave",
		endTime: models.DrawDownTime{Minutes: 3, Seconds: 0},
	},
	{
		name:             "Kenya AA",
		origin:           "Kenya",
		roaster:          "Intelligentsia",
		roastLevel:       "light medium",
		processingMethod: "washed",
		tastingNotes:     [5]string{"blackcurrant", "grapefruit", "wine", "tomato", "blackberry"},
		tastingTraits: models.TastingTraits{
			BerryIntensity:        9,
			StonefruitIntensity:   5,
			RoastIntensity:        3,
			CitrusFruitsIntensity: 8,
			Bitterness:            2,
			Florality:             6,
			Spice:                 3,
			Sweetness:             6,
			AromaticIntensity:     8,
			Savory:                5,
			Body:                  6,
			Cleanliness:           9,
		},
		rating:  9,
		recipe:  []string{"22g coffee", "350ml water", "94°C", "Chemex"},
		dripper: "Chemex",
		endTime: models.DrawDownTime{Minutes: 4, Seconds: 15},
	},
	{
		name:             "Guatemala Huehuetenango",
		origin:           "Guatemala",
		roaster:          "Stumptown",
		roastLevel:       "medium",
		processingMethod: "washed",
		tastingNotes:     [5]string{"apple", "caramel", "cocoa", "almond", "honey"},
		tastingTraits: models.TastingTraits{
			BerryIntensity:        3,
			StonefruitIntensity:   6,
			RoastIntensity:        4,
			CitrusFruitsIntensity: 4,
			Bitterness:            2,
			Florality:             4,
			Spice:                 3,
			Sweetness:             8,
			AromaticIntensity:     7,
			Savory:                3,
			Body:                  6,
			Cleanliness:           8,
		},
		rating:  8,
		recipe:  []string{"19g coffee", "310ml water", "92°C", "V60 pour over"},
		dripper: "Hario V60",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 50},
	},
	{
		name:             "Sumatra Mandheling",
		origin:           "Indonesia",
		roaster:          "Peet's Coffee",
		roastLevel:       "dark",
		processingMethod: "natural",
		tastingNotes:     [5]string{"earth", "tobacco", "dark chocolate", "spice", "cedar"},
		tastingTraits: models.TastingTraits{
			BerryIntensity:        1,
			StonefruitIntensity:   2,
			RoastIntensity:        9,
			CitrusFruitsIntensity: 1,
			Bitterness:            6,
			Florality:             1,
			Spice:                 7,
			Sweetness:             4,
			AromaticIntensity:     8,
			Savory:                8,
			Body:                  9,
			Cleanliness:           6,
		},
		rating:  7,
		recipe:  []string{"17g coffee", "280ml water", "88°C", "French Press"},
		dripper: "French Press",
		endTime: models.DrawDownTime{Minutes: 4, Seconds: 0},
	},
	{
		name:             "Costa Rica Tarrazu",
		origin:           "Costa Rica",
		roaster:          "Verve",
		roastLevel:       "light",
		processingMethod: "honey",
		tastingNotes:     [5]string{"peach", "honey", "vanilla", "lemon", "floral"},
		tastingTraits: models.TastingTraits{
			BerryIntensity:        4,
			StonefruitIntensity:   8,
			RoastIntensity:        2,
			CitrusFruitsIntensity: 6,
			Bitterness:            1,
			Florality:             7,
			Spice:                 2,
			Sweetness:             9,
			AromaticIntensity:     8,
			Savory:                2,
			Body:                  5,
			Cleanliness:           8,
		},
		rating:  9,
		recipe:  []string{"21g coffee", "330ml water", "94°C", "Clever Dripper"},
		dripper: "Clever Dripper",
		endTime: models.DrawDownTime{Minutes: 3, Seconds: 30},
	},
}

func main() {
	// Parse command-line flags
	storageType := flag.String("storage", "mysql", "Storage type: memory or mysql")
	mysqlHost := flag.String("mysql-host", "localhost:3306", "MySQL host")
	mysqlUser := flag.String("mysql-user", "coffee_user", "MySQL user")
	mysqlPassword := flag.String("mysql-password", "coffee_pass123", "MySQL password")
	mysqlDB := flag.String("mysql-db", "coffee_log", "MySQL database name")
	count := flag.Int("count", 0, "Number of entries to create (0 = all sample data)")
	flag.Parse()

	// Initialize storage
	var store storage.CoffeeStorage
	var err error

	switch *storageType {
	case "mysql":
		store, err = storage.NewMySQLStorage(*mysqlHost, *mysqlUser, *mysqlPassword, *mysqlDB)
		if err != nil {
			log.Fatalf("Failed to initialize MySQL storage: %v", err)
		}
		fmt.Println("✅ Connected to MySQL storage")

		// Close MySQL connection on exit
		if mysqlStore, ok := store.(*storage.MySQLStorage); ok {
			defer mysqlStore.Close()
		}
	case "memory":
		store = storage.NewMemoryStorage()
		fmt.Println("✅ Using in-memory storage")
	default:
		log.Fatalf("Invalid storage type: %s", *storageType)
	}

	// Determine how many entries to create
	numEntries := len(sampleCoffees)
	if *count > 0 && *count < numEntries {
		numEntries = *count
	}

	fmt.Printf("\n📦 Creating %d test coffee entries...\n\n", numEntries)

	// Create coffee entries
	for i := 0; i < numEntries; i++ {
		sample := sampleCoffees[i]
		now := time.Now()

		coffee := models.Coffee{
			ID:               uuid.New().String(),
			Name:             sample.name,
			Origin:           sample.origin,
			Roaster:          sample.roaster,
			RoastLevel:       sample.roastLevel,
			ProcessingMethod: sample.processingMethod,
			TastingNotes:     sample.tastingNotes,
			TastingTraits:    sample.tastingTraits,
			Rating:           sample.rating,
			Recipe:           sample.recipe,
			Dripper:          sample.dripper,
			EndTime:          sample.endTime,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		// Validate before saving
		if err := coffee.Validate(); err != nil {
			log.Printf("❌ Validation error for %s: %v", coffee.Name, err)
			continue
		}

		// Save to storage
		if err := store.Save(coffee); err != nil {
			log.Printf("❌ Failed to save %s: %v", coffee.Name, err)
			continue
		}

		fmt.Printf("✅ Created: %s (ID: %s)\n", coffee.Name, coffee.ID)
	}

	fmt.Printf("\n🎉 Successfully created %d test entries!\n", numEntries)
	fmt.Println("\n📊 To view all entries:")
	fmt.Println("   curl http://localhost:8080/coffees")
	fmt.Println("\n🚀 To start the server:")
	if *storageType == "mysql" {
		fmt.Printf("   go run main.go -storage=mysql -mysql-user=%s -mysql-password=%s -mysql-db=%s\n",
			*mysqlUser, *mysqlPassword, *mysqlDB)
	} else {
		fmt.Println("   go run main.go")
	}
}