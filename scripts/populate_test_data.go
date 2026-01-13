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

// Sample coffee data for testing - bean info only
var sampleCoffees = []struct {
	name             string
	origin           string
	roaster          string
	variety          string
	roastLevel       string
	processingMethod string
}{
	{
		name:             "Ethiopian Yirgacheffe",
		origin:           "Ethiopia",
		roaster:          "Blue Bottle",
		variety:          "Heirloom",
		roastLevel:       "light",
		processingMethod: "washed",
	},
	{
		name:             "Colombian Supremo",
		origin:           "Colombia",
		roaster:          "Counter Culture",
		variety:          "Caturra",
		roastLevel:       "medium",
		processingMethod: "washed",
	},
	{
		name:             "Kenya AA",
		origin:           "Kenya",
		roaster:          "Intelligentsia",
		variety:          "SL28",
		roastLevel:       "light medium",
		processingMethod: "washed",
	},
	{
		name:             "Sumatra Mandheling",
		origin:           "Indonesia",
		roaster:          "Stumptown",
		variety:          "Typica",
		roastLevel:       "dark",
		processingMethod: "wet hulled",
	},
	{
		name:             "Guatemala Antigua",
		origin:           "Guatemala",
		roaster:          "Verve",
		variety:          "Bourbon",
		roastLevel:       "medium",
		processingMethod: "washed",
	},
	{
		name:             "Panama Geisha",
		origin:           "Panama",
		roaster:          "Onyx",
		variety:          "Geisha",
		roastLevel:       "light",
		processingMethod: "washed",
	},
	{
		name:             "Brazil Natural",
		origin:           "Brazil",
		roaster:          "Heart",
		variety:          "Yellow Bourbon",
		roastLevel:       "medium dark",
		processingMethod: "natural",
	},
	{
		name:             "Costa Rica Honey",
		origin:           "Costa Rica",
		roaster:          "Sey",
		variety:          "Catuai",
		roastLevel:       "light medium",
		processingMethod: "honey",
	},
}

// Sample brew data - tasting/evaluation info
var sampleBrews = []struct {
	coffeeIndex   int
	tastingNotes  [5]string
	tastingTraits models.TastingTraits
	rating        int
	recipe        []string
	dripper       string
	endTime       models.DrawDownTime
}{
	// Ethiopian Yirgacheffe brews
	{
		coffeeIndex:  0,
		tastingNotes: [5]string{"blueberry", "jasmine", "honey", "citrus", "tea"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 8, StonefruitIntensity: 3, RoastIntensity: 2,
			CitrusFruitsIntensity: 7, Bitterness: 1, Florality: 9,
			Spice: 2, Sweetness: 8, AromaticIntensity: 9,
			Savory: 1, Body: 5, Cleanliness: 9,
		},
		rating:  9,
		recipe:  []string{"20g coffee", "320ml water", "95°C", "V60 pour over"},
		dripper: "Hario V60",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 45},
	},
	{
		coffeeIndex:  0,
		tastingNotes: [5]string{"blueberry", "floral", "bergamot", "lemon", "honey"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 9, StonefruitIntensity: 2, RoastIntensity: 2,
			CitrusFruitsIntensity: 8, Bitterness: 1, Florality: 8,
			Spice: 1, Sweetness: 9, AromaticIntensity: 8,
			Savory: 1, Body: 4, Cleanliness: 9,
		},
		rating:  9,
		recipe:  []string{"18g coffee", "300ml water", "94°C", "V60 pour over"},
		dripper: "Hario V60",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 30},
	},
	// Colombian Supremo brews
	{
		coffeeIndex:  1,
		tastingNotes: [5]string{"chocolate", "caramel", "nuts", "orange", "brown sugar"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 2, StonefruitIntensity: 4, RoastIntensity: 5,
			CitrusFruitsIntensity: 5, Bitterness: 3, Florality: 3,
			Spice: 2, Sweetness: 7, AromaticIntensity: 6,
			Savory: 4, Body: 7, Cleanliness: 8,
		},
		rating:  8,
		recipe:  []string{"18g coffee", "300ml water", "93°C", "Kalita Wave"},
		dripper: "Kalita Wave",
		endTime: models.DrawDownTime{Minutes: 3, Seconds: 0},
	},
	// Kenya AA brews
	{
		coffeeIndex:  2,
		tastingNotes: [5]string{"blackcurrant", "grapefruit", "wine", "tomato", "blackberry"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 9, StonefruitIntensity: 5, RoastIntensity: 3,
			CitrusFruitsIntensity: 8, Bitterness: 2, Florality: 6,
			Spice: 3, Sweetness: 6, AromaticIntensity: 8,
			Savory: 5, Body: 6, Cleanliness: 9,
		},
		rating:  8,
		recipe:  []string{"20g coffee", "340ml water", "96°C", "V60 pour over"},
		dripper: "Hario V60",
		endTime: models.DrawDownTime{Minutes: 3, Seconds: 15},
	},
	// Sumatra Mandheling brews
	{
		coffeeIndex:  3,
		tastingNotes: [5]string{"earthy", "tobacco", "dark chocolate", "cedar", "mushroom"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 1, StonefruitIntensity: 2, RoastIntensity: 8,
			CitrusFruitsIntensity: 1, Bitterness: 6, Florality: 1,
			Spice: 6, Sweetness: 3, AromaticIntensity: 7,
			Savory: 8, Body: 9, Cleanliness: 5,
		},
		rating:  7,
		recipe:  []string{"22g coffee", "350ml water", "91°C", "French Press"},
		dripper: "French Press",
		endTime: models.DrawDownTime{Minutes: 4, Seconds: 0},
	},
	// Guatemala Antigua brews
	{
		coffeeIndex:  4,
		tastingNotes: [5]string{"cocoa", "apple", "spice", "caramel", "almond"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 3, StonefruitIntensity: 6, RoastIntensity: 5,
			CitrusFruitsIntensity: 4, Bitterness: 3, Florality: 4,
			Spice: 5, Sweetness: 6, AromaticIntensity: 6,
			Savory: 3, Body: 7, Cleanliness: 8,
		},
		rating:  8,
		recipe:  []string{"19g coffee", "310ml water", "93°C", "Origami"},
		dripper: "Origami",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 50},
	},
	// Panama Geisha brews
	{
		coffeeIndex:  5,
		tastingNotes: [5]string{"jasmine", "bergamot", "peach", "mango", "rose"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 4, StonefruitIntensity: 8, RoastIntensity: 2,
			CitrusFruitsIntensity: 6, Bitterness: 1, Florality: 10,
			Spice: 2, Sweetness: 9, AromaticIntensity: 10,
			Savory: 1, Body: 4, Cleanliness: 10,
		},
		rating:  10,
		recipe:  []string{"15g coffee", "250ml water", "94°C", "V60 pour over"},
		dripper: "Hario V60",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 20},
	},
	// Brazil Natural brews
	{
		coffeeIndex:  6,
		tastingNotes: [5]string{"peanut", "milk chocolate", "hazelnut", "dried fruit", "whiskey"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 3, StonefruitIntensity: 3, RoastIntensity: 6,
			CitrusFruitsIntensity: 2, Bitterness: 4, Florality: 2,
			Spice: 3, Sweetness: 6, AromaticIntensity: 5,
			Savory: 5, Body: 8, Cleanliness: 6,
		},
		rating:  7,
		recipe:  []string{"20g coffee", "320ml water", "92°C", "Chemex"},
		dripper: "Chemex",
		endTime: models.DrawDownTime{Minutes: 3, Seconds: 30},
	},
	// Costa Rica Honey brews
	{
		coffeeIndex:  7,
		tastingNotes: [5]string{"honey", "plum", "orange zest", "brown sugar", "stone fruit"},
		tastingTraits: models.TastingTraits{
			BerryIntensity: 4, StonefruitIntensity: 7, RoastIntensity: 4,
			CitrusFruitsIntensity: 6, Bitterness: 2, Florality: 5,
			Spice: 3, Sweetness: 8, AromaticIntensity: 7,
			Savory: 2, Body: 6, Cleanliness: 8,
		},
		rating:  8,
		recipe:  []string{"18g coffee", "290ml water", "93°C", "Kalita Wave"},
		dripper: "Kalita Wave",
		endTime: models.DrawDownTime{Minutes: 2, Seconds: 55},
	},
}

func main() {
	storageType := flag.String("storage", "memory", "Storage type: memory or mysql")
	mysqlUser := flag.String("mysql-user", "coffee_user", "MySQL username")
	mysqlPassword := flag.String("mysql-password", "coffee_pass", "MySQL password")
	mysqlDB := flag.String("mysql-db", "coffee_dex", "MySQL database name")
	mysqlHost := flag.String("mysql-host", "localhost", "MySQL host")
	mysqlPort := flag.String("mysql-port", "3306", "MySQL port")
	flag.Parse()

	var coffeeStore storage.CoffeeStorage
	var brewStore storage.BrewStorage

	switch *storageType {
	case "mysql":
		hostWithPort := fmt.Sprintf("%s:%s", *mysqlHost, *mysqlPort)
		mysqlStorage, err := storage.NewMySQLStorage(hostWithPort, *mysqlUser, *mysqlPassword, *mysqlDB)
		if err != nil {
			log.Fatalf("Failed to connect to MySQL: %v", err)
		}

		coffeeStore = mysqlStorage
		brewStore = storage.NewMySQLBrewStorage(mysqlStorage.DB())

		fmt.Println("Connected to MySQL database")
	default:
		log.Fatal("Only MySQL storage is supported for test data population")
	}

	fmt.Println("\nCreating test coffees and brews...")
	fmt.Println("==================================")

	// Store created coffee IDs for brew references
	coffeeIDs := make([]string, len(sampleCoffees))

	// Create coffees first
	for i, sample := range sampleCoffees {
		now := time.Now()

		coffee := models.Coffee{
			ID:               uuid.New().String(),
			Name:             sample.name,
			Origin:           sample.origin,
			Roaster:          sample.roaster,
			Variety:          sample.variety,
			RoastLevel:       sample.roastLevel,
			ProcessingMethod: sample.processingMethod,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := coffee.Validate(); err != nil {
			log.Printf("Validation error for %s: %v", coffee.Name, err)
			continue
		}

		if err := coffeeStore.Save(coffee); err != nil {
			log.Printf("Failed to save %s: %v", coffee.Name, err)
			continue
		}

		coffeeIDs[i] = coffee.ID
		fmt.Printf("Created coffee: %s (ID: %s)\n", coffee.Name, coffee.ID)
	}

	// Create brews
	brewCount := 0
	for _, brewData := range sampleBrews {
		coffeeID := coffeeIDs[brewData.coffeeIndex]
		if coffeeID == "" {
			continue
		}

		brew := models.Brew{
			ID:            uuid.New().String(),
			CoffeeID:      coffeeID,
			TastingNotes:  brewData.tastingNotes,
			TastingTraits: brewData.tastingTraits,
			Rating:        brewData.rating,
			Recipe:        brewData.recipe,
			Dripper:       brewData.dripper,
			EndTime:       brewData.endTime,
			CreatedAt:     time.Now(),
		}

		if err := brew.Validate(); err != nil {
			log.Printf("Validation error for brew: %v", err)
			continue
		}

		if err := brewStore.Save(brew); err != nil {
			log.Printf("Failed to save brew: %v", err)
			continue
		}

		brewCount++
		fmt.Printf("  Created brew for %s (ID: %s)\n", sampleCoffees[brewData.coffeeIndex].name, brew.ID)
	}

	fmt.Printf("\nSuccessfully created %d coffees and %d brews!\n", len(sampleCoffees), brewCount)
	fmt.Println("\nTo view all entries:")
	fmt.Println("   curl http://localhost:8080/coffees")
	fmt.Println("   curl http://localhost:8080/brews")
	fmt.Println("\nTo start the server:")
	if *storageType == "mysql" {
		fmt.Printf("   go run main.go -storage=mysql -mysql-user=%s -mysql-password=%s -mysql-db=%s\n",
			*mysqlUser, *mysqlPassword, *mysqlDB)
	}
}
