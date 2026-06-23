package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DateOnly is a custom type that handles JSON dates in "YYYY-MM-DD" format
type DateOnly time.Time

// UnmarshalJSON parses dates in "YYYY-MM-DD" format (from HTML date inputs)
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	// Try parsing as simple date first (YYYY-MM-DD)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		// Fall back to RFC3339
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("invalid date format: %s", s)
		}
	}
	*d = DateOnly(t)
	return nil
}

// MarshalJSON outputs the date in "YYYY-MM-DD" format
func (d DateOnly) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(t.Format("2006-01-02"))
}

// Time returns the underlying time.Time value
func (d DateOnly) Time() time.Time {
	return time.Time(d)
}

// IsZero reports whether the date is zero
func (d DateOnly) IsZero() bool {
	return time.Time(d).IsZero()
}

// DrawDownTime represents brew time in minutes and seconds
type DrawDownTime struct {
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

// TastingTraits represents the 12 flavor profile traits (0-10 scale)
type TastingTraits struct {
	BerryIntensity        int `json:"berry_intensity"`
	StonefruitIntensity   int `json:"stonefruit_intensity"`
	RoastIntensity        int `json:"roast_intensity"`
	CitrusFruitsIntensity int `json:"citrus_fruits_intensity"`
	Bitterness            int `json:"bitterness"`
	Florality             int `json:"florality"`
	Spice                 int `json:"spice"`
	Sweetness             int `json:"sweetness"`
	AromaticIntensity     int `json:"aromatic_intensity"`
	Savory                int `json:"savory"`
	Body                  int `json:"body"`
	Cleanliness           int `json:"cleanliness"`
}

// Coffee represents a coffee bean entry (no tasting data - that's in Brew)
type Coffee struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Origin           string    `json:"origin"`
	Roaster          string    `json:"roaster"`
	Variety          string    `json:"variety"`
	RoastLevel       string    `json:"roast_level"`
	ProcessingMethod string    `json:"processing_method"`
	RoastDate        *DateOnly `json:"roast_date,omitempty"`
	IsFinished       bool       `json:"is_finished"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// DaysOpen returns the number of days since the bag was opened (created).
func (c *Coffee) DaysOpen() int {
	return int(time.Since(c.CreatedAt).Hours() / 24)
}

// DaysOffRoast calculates days since roasting. Returns -1 if no roast date set.
func (c *Coffee) DaysOffRoast() int {
	if c.RoastDate == nil || c.RoastDate.IsZero() {
		return -1
	}
	return int(time.Since(c.RoastDate.Time()).Hours() / 24)
}

// DaysOffRoastAt calculates days off roast at a specific time (e.g., when a brew was made)
func (c *Coffee) DaysOffRoastAt(t time.Time) int {
	if c.RoastDate == nil || c.RoastDate.IsZero() {
		return -1
	}
	return int(t.Sub(c.RoastDate.Time()).Hours() / 24)
}

// CoffeeWithBrewStats includes brew statistics for UI display
type CoffeeWithBrewStats struct {
	Coffee
	BrewCount          int     `json:"brew_count"`
	AverageRating      float64 `json:"average_rating"`
	CanGeneratePokemon bool    `json:"can_generate_pokemon"`
	HasPokemon         bool    `json:"has_pokemon"`
	DaysOffRoast       int     `json:"days_off_roast"` // -1 if no roast date
}

func (t *TastingTraits) Validate() error {
	traits := []struct {
		name  string
		value int
	}{
		{"berry_intensity", t.BerryIntensity},
		{"stonefruit_intensity", t.StonefruitIntensity},
		{"roast_intensity", t.RoastIntensity},
		{"citrus_fruits_intensity", t.CitrusFruitsIntensity},
		{"bitterness", t.Bitterness},
		{"florality", t.Florality},
		{"spice", t.Spice},
		{"sweetness", t.Sweetness},
		{"aromatic_intensity", t.AromaticIntensity},
		{"savory", t.Savory},
		{"body", t.Body},
		{"cleanliness", t.Cleanliness},
	}

	for _, trait := range traits {
		// -1 means "not scored" (trait was skipped), 0-10 is the valid range
		if trait.value != -1 && (trait.value < 0 || trait.value > 10) {
			return fmt.Errorf("%s must be -1 (not scored) or between 0 and 10, got %d", trait.name, trait.value)
		}
	}

	return nil
}

func (c *Coffee) ValidateProcessingMethod() error {
	c.ProcessingMethod = strings.ToLower(c.ProcessingMethod)
	validMethods := []string{"washed", "natural", "honey", "coferment", "experimental"}
	for _, method := range validMethods {
		if c.ProcessingMethod == method {
			return nil
		}
	}
	return fmt.Errorf("invalid processing method: %s", c.ProcessingMethod)
}

func (c *Coffee) ValidateRoastLevel() error {
	c.RoastLevel = strings.ToLower(c.RoastLevel)
	validLevels := []string{"light", "medium", "dark", "light medium", "medium dark", "unclear"}
	for _, level := range validLevels {
		if c.RoastLevel == level {
			return nil
		}
	}
	return fmt.Errorf("invalid roast level: %s", c.RoastLevel)
}

// Validate checks if the Coffee data is valid
func (c *Coffee) Validate() error {
	// Only name is required
	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Validate roast level if provided
	if c.RoastLevel != "" {
		if err := c.ValidateRoastLevel(); err != nil {
			return err
		}
	}

	// Validate processing method if provided
	if c.ProcessingMethod != "" {
		if err := c.ValidateProcessingMethod(); err != nil {
			return err
		}
	}

	return nil
}
