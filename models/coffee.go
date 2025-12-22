package models

import (
	"fmt"
	"strings"
	"time"
)

// Coffee represents a coffee tasting entry

type DrawDownTime struct {
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type TastingTraits struct {
	BerryIntensity       int `json:"berry_intensity"`
	StonefruitIntensity  int `json:"stonefruit_intensity"`
	RoastIntensity       int `json:"roast_intensity"`
	CitrusFruitsIntensity int `json:"citrus_fruits_intensity"`
	Bitterness           int `json:"bitterness"`
	Florality            int `json:"florality"`
	Spice                int `json:"spice"`
	Sweetness            int `json:"sweetness"`
	AromaticIntensity    int `json:"aromatic_intensity"`
	Savory               int `json:"savory"`
	Body                 int `json:"body"`
	Cleanliness          int `json:"cleanliness"`
}

type Coffee struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Origin string `json:"origin"`
	Roaster string `json:"roaster"`
	Variety string `json:"variety"`
	RoastLevel string `json:"roast_level"`
	ProcessingMethod string `json:"processing_method"`
	TastingNotes [5]string `json:"tasting_notes"`
	TastingTraits TastingTraits `json:"tasting_traits"`
	Rating int `json:"rating"`
	Recipe []string `json:"recipe"`
	Dripper string `json:"dripper"`
	EndTime DrawDownTime `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		if trait.value < 0 || trait.value > 10 {
			return fmt.Errorf("%s must be between 0 and 10, got %d", trait.name, trait.value)
		}
	}
	
	return nil
}

func (c *Coffee) ValidateProcessingMethod() error {
	c.ProcessingMethod = strings.ToLower(c.ProcessingMethod)
	validMethods := []string{"washed", "natural", "honey", "coferment", "experimental"}
	for method := range validMethods {
		if c.ProcessingMethod == validMethods[method] {
			return nil
		}
	}
	return fmt.Errorf("invalid processing method: %s", c.ProcessingMethod)
}

func (c *Coffee) ValidateRoastLevel() error {
	c.RoastLevel = strings.ToLower(c.RoastLevel)
	validLevels := []string{"light", "medium", "dark", "light medium", "medium dark", "unclear"}
	for level := range validLevels {
		if c.RoastLevel == validLevels[level] {
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
	
	// Validate rating if provided
	if c.Rating < 0 || c.Rating > 10 {
		return fmt.Errorf("ratings must be out of 10")
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
	
	// Tasting notes are optional - just check length if provided
	if len(c.TastingNotes) > 5 {
		return fmt.Errorf("tasting notes maximum length is 5")
	}
	
	// Validate draw down time if provided
	if c.EndTime.Minutes < 0 || c.EndTime.Seconds < 0 || c.EndTime.Seconds >= 60 {
		return fmt.Errorf("invalid draw down time")
	}
	
	// Validate tasting traits - allow default values
	if err := c.TastingTraits.Validate(); err != nil {
		return err
	}
	
	return nil
}