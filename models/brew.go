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
	DaysOffRoast  int           `json:"days_off_roast"` // Calculated: days between roast date and brew date (-1 if no roast date)
	IsLearning    bool          `json:"is_learning"`    // If true, excluded from trait/rating aggregation
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
	IsFinished         bool `json:"is_finished"`
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

// averageTrait computes the average of a single trait across brews, skipping -1 (not scored)
func averageTrait(brews []Brew, getter func(TastingTraits) int) int {
	sum, count := 0, 0
	for _, b := range brews {
		v := getter(b.TastingTraits)
		if v >= 0 {
			sum += v
			count++
		}
	}
	if count == 0 {
		return -1
	}
	return sum / count
}

// AverageTraits computes the average TastingTraits from a slice of brews.
// Traits with value -1 (not scored) are excluded from the average.
// If no brews scored a given trait, the average is -1.
func AverageTraits(brews []Brew) TastingTraits {
	if len(brews) == 0 {
		return TastingTraits{}
	}

	return TastingTraits{
		BerryIntensity:        averageTrait(brews, func(t TastingTraits) int { return t.BerryIntensity }),
		StonefruitIntensity:   averageTrait(brews, func(t TastingTraits) int { return t.StonefruitIntensity }),
		RoastIntensity:        averageTrait(brews, func(t TastingTraits) int { return t.RoastIntensity }),
		CitrusFruitsIntensity: averageTrait(brews, func(t TastingTraits) int { return t.CitrusFruitsIntensity }),
		Bitterness:            averageTrait(brews, func(t TastingTraits) int { return t.Bitterness }),
		Florality:             averageTrait(brews, func(t TastingTraits) int { return t.Florality }),
		Spice:                 averageTrait(brews, func(t TastingTraits) int { return t.Spice }),
		Sweetness:             averageTrait(brews, func(t TastingTraits) int { return t.Sweetness }),
		AromaticIntensity:     averageTrait(brews, func(t TastingTraits) int { return t.AromaticIntensity }),
		Savory:                averageTrait(brews, func(t TastingTraits) int { return t.Savory }),
		Body:                  averageTrait(brews, func(t TastingTraits) int { return t.Body }),
		Cleanliness:           averageTrait(brews, func(t TastingTraits) int { return t.Cleanliness }),
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
