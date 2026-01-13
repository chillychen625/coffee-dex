package models

import (
	"fmt"
	"time"
)

// Brew represents a single tasting/brewing session of a coffee
type Brew struct {
	ID            string        `json:"id"`
	CoffeeID      string        `json:"coffee_id"`
	TastingNotes  [5]string     `json:"tasting_notes"`
	TastingTraits TastingTraits `json:"tasting_traits"`
	Rating        int           `json:"rating"`
	Recipe        []string      `json:"recipe"`
	Dripper       string        `json:"dripper"`
	EndTime       DrawDownTime  `json:"end_time"`
	CreatedAt     time.Time     `json:"created_at"`
}

// BrewWithCoffee includes the coffee information along with the brew
type BrewWithCoffee struct {
	Brew
	CoffeeName   string `json:"coffee_name"`
	CoffeeOrigin string `json:"coffee_origin"`
}

// AggregatedBrewData represents aggregated data from multiple brews of a coffee
type AggregatedBrewData struct {
	CoffeeID        string        `json:"coffee_id"`
	BrewCount       int           `json:"brew_count"`
	AverageRating   float64       `json:"average_rating"`
	AverageTraits   TastingTraits `json:"average_traits"`
	CombinedNotes   []string      `json:"combined_notes"`
	MostUsedDripper string        `json:"most_used_dripper"`
}

// BrewProgress represents the progress toward Pokemon generation
type BrewProgress struct {
	Count              int  `json:"count"`
	Required           int  `json:"required"`
	CanGeneratePokemon bool `json:"can_generate_pokemon"`
	HasPokemon         bool `json:"has_pokemon"`
}

// RequiredBrewsForPokemon is the number of brews needed to unlock Pokemon generation
const RequiredBrewsForPokemon = 5

// Validate checks if the Brew data is valid
func (b *Brew) Validate() error {
	if b.CoffeeID == "" {
		return fmt.Errorf("coffee_id cannot be empty")
	}

	// Validate rating
	if b.Rating < 0 || b.Rating > 10 {
		return fmt.Errorf("rating must be between 0 and 10, got %d", b.Rating)
	}

	// Validate draw down time if provided
	if b.EndTime.Minutes < 0 || b.EndTime.Seconds < 0 || b.EndTime.Seconds >= 60 {
		return fmt.Errorf("invalid draw down time")
	}

	// Validate tasting traits
	if err := b.TastingTraits.Validate(); err != nil {
		return err
	}

	return nil
}

// AverageTraits computes the average TastingTraits from a slice of brews
func AverageTraits(brews []Brew) TastingTraits {
	if len(brews) == 0 {
		return TastingTraits{}
	}

	var sum TastingTraits
	for _, b := range brews {
		sum.BerryIntensity += b.TastingTraits.BerryIntensity
		sum.StonefruitIntensity += b.TastingTraits.StonefruitIntensity
		sum.RoastIntensity += b.TastingTraits.RoastIntensity
		sum.CitrusFruitsIntensity += b.TastingTraits.CitrusFruitsIntensity
		sum.Bitterness += b.TastingTraits.Bitterness
		sum.Florality += b.TastingTraits.Florality
		sum.Spice += b.TastingTraits.Spice
		sum.Sweetness += b.TastingTraits.Sweetness
		sum.AromaticIntensity += b.TastingTraits.AromaticIntensity
		sum.Savory += b.TastingTraits.Savory
		sum.Body += b.TastingTraits.Body
		sum.Cleanliness += b.TastingTraits.Cleanliness
	}

	count := len(brews)
	return TastingTraits{
		BerryIntensity:        sum.BerryIntensity / count,
		StonefruitIntensity:   sum.StonefruitIntensity / count,
		RoastIntensity:        sum.RoastIntensity / count,
		CitrusFruitsIntensity: sum.CitrusFruitsIntensity / count,
		Bitterness:            sum.Bitterness / count,
		Florality:             sum.Florality / count,
		Spice:                 sum.Spice / count,
		Sweetness:             sum.Sweetness / count,
		AromaticIntensity:     sum.AromaticIntensity / count,
		Savory:                sum.Savory / count,
		Body:                  sum.Body / count,
		Cleanliness:           sum.Cleanliness / count,
	}
}

// CombineNotes collects all unique tasting notes from brews
func CombineNotes(brews []Brew) []string {
	noteSet := make(map[string]bool)
	for _, b := range brews {
		for _, note := range b.TastingNotes {
			if note != "" {
				noteSet[note] = true
			}
		}
	}

	notes := make([]string, 0, len(noteSet))
	for note := range noteSet {
		notes = append(notes, note)
	}
	return notes
}

// MostUsedDripper finds the most frequently used dripper from brews
func MostUsedDripper(brews []Brew) string {
	counts := make(map[string]int)
	for _, b := range brews {
		if b.Dripper != "" {
			counts[b.Dripper]++
		}
	}

	var maxDripper string
	var maxCount int
	for dripper, count := range counts {
		if count > maxCount {
			maxCount = count
			maxDripper = dripper
		}
	}
	return maxDripper
}

// AverageRating computes the average rating from brews
func AverageRating(brews []Brew) float64 {
	if len(brews) == 0 {
		return 0
	}

	var sum float64
	for _, b := range brews {
		sum += float64(b.Rating)
	}
	return sum / float64(len(brews))
}
