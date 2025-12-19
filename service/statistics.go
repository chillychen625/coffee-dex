package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"math"
	"sort"
)

// StatisticsService handles analytics and statistics calculations
type StatisticsService struct {
	coffeeStorage  storage.CoffeeStorage
	pokemonStorage storage.PokemonStorage
	mapper         *PokemonMapper
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(
	coffeeStorage storage.CoffeeStorage,
	pokemonStorage storage.PokemonStorage,
) *StatisticsService {
	return &StatisticsService{
		coffeeStorage:  coffeeStorage,
		pokemonStorage: pokemonStorage,
		mapper:         NewPokemonMapper(),
	}
}

// Statistics represents overall coffee collection statistics
type Statistics struct {
	// Basic counts
	TotalCoffees      int                       `json:"total_coffees"`
	TotalPokemon      int                       `json:"total_pokemon"`
	CompletionPercent float64                   `json:"completion_percent"`
	
	// Ratings
	AverageRating     float64                   `json:"average_rating"`
	HighestRated      *CoffeeRatingSummary      `json:"highest_rated"`
	LowestRated       *CoffeeRatingSummary      `json:"lowest_rated"`
	
	// Type distribution
	TypeDistribution  map[string]int            `json:"type_distribution"`
	MostCommonType    string                    `json:"most_common_type"`
	
	// Origin statistics
	OriginDistribution map[string]int           `json:"origin_distribution"`
	TopOrigins        []OriginStat              `json:"top_origins"`
	
	// Processing methods
	ProcessingStats   map[string]ProcessingStat `json:"processing_stats"`
	
	// Roast levels
	RoastDistribution map[string]int            `json:"roast_distribution"`
	
	// Trait analysis
	TraitAverages     models.TastingTraits      `json:"trait_averages"`
	TraitRanges       TraitRanges               `json:"trait_ranges"`
	
	// Brewer analysis
	BrewerStats       map[string]BrewerStat     `json:"brewer_stats"`
	
	// Confidence metrics
	AverageConfidence float64                   `json:"average_confidence"`
	HighConfidencePairings int                  `json:"high_confidence_pairings"` // >= 0.8
}

// CoffeeRatingSummary represents a summary of a coffee for rating display
type CoffeeRatingSummary struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Origin       string  `json:"origin"`
	Rating       int     `json:"rating"`
	PokemonName  string  `json:"pokemon_name,omitempty"`
}

// OriginStat represents statistics for a coffee origin
type OriginStat struct {
	Origin        string  `json:"origin"`
	Count         int     `json:"count"`
	AverageRating float64 `json:"average_rating"`
}

// ProcessingStat represents statistics for a processing method
type ProcessingStat struct {
	Count         int     `json:"count"`
	AverageRating float64 `json:"average_rating"`
	CommonTypes   []string `json:"common_types"`
}

// BrewerStat represents statistics for a brewing device
type BrewerStat struct {
	Count         int     `json:"count"`
	AverageRating float64 `json:"average_rating"`
	AvgBrewTime   float64 `json:"avg_brew_time_seconds"`
}

// TraitRanges represents min/max ranges for tasting traits
type TraitRanges struct {
	BerryRange      Range `json:"berry_range"`
	StonefruitRange Range `json:"stonefruit_range"`
	RoastRange      Range `json:"roast_range"`
	CitrusRange     Range `json:"citrus_range"`
	BitternessRange Range `json:"bitterness_range"`
	FloralityRange  Range `json:"florality_range"`
	SpiceRange      Range `json:"spice_range"`
	SweetnessRange  Range `json:"sweetness_range"`
	AromaticRange   Range `json:"aromatic_range"`
	SavoryRange     Range `json:"savory_range"`
	BodyRange       Range `json:"body_range"`
	CleanlinessRange Range `json:"cleanliness_range"`
}

// Range represents a min/max range
type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// CalculateStatistics computes all statistics from the database
func (s *StatisticsService) CalculateStatistics() (*Statistics, error) {
	// Get all coffees and pokemon mappings
	coffees, err := s.coffeeStorage.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get coffees: %w", err)
	}
	
	pokemonMappings, err := s.pokemonStorage.GetAllCoffeePokemon()
	if err != nil {
		return nil, fmt.Errorf("failed to get pokemon mappings: %w", err)
	}
	
	stats := &Statistics{
		TotalCoffees:      len(coffees),
		TotalPokemon:      len(pokemonMappings),
		CompletionPercent: float64(len(pokemonMappings)) / 151.0 * 100.0,
		TypeDistribution:  make(map[string]int),
		OriginDistribution: make(map[string]int),
		ProcessingStats:   make(map[string]ProcessingStat),
		RoastDistribution: make(map[string]int),
		BrewerStats:       make(map[string]BrewerStat),
	}
	
	// Calculate statistics
	s.calculateRatingStats(coffees, pokemonMappings, stats)
	s.calculateTypeDistribution(coffees, stats)
	s.calculateOriginStats(coffees, stats)
	s.calculateProcessingStats(coffees, stats)
	s.calculateRoastDistribution(coffees, stats)
	s.calculateTraitAverages(coffees, stats)
	s.calculateBrewerStats(coffees, stats)
	s.calculateConfidenceMetrics(pokemonMappings, stats)
	
	return stats, nil
}

// calculateRatingStats calculates rating-based statistics
func (s *StatisticsService) calculateRatingStats(coffees []models.Coffee, mappings []models.CoffeePokemon, stats *Statistics) {
	if len(coffees) == 0 {
		return
	}
	
	totalRating := 0
	var highest, lowest *models.Coffee
	
	for i := range coffees {
		coffee := &coffees[i]
		totalRating += coffee.Rating
		
		if highest == nil || coffee.Rating > highest.Rating {
			highest = coffee
		}
		if lowest == nil || coffee.Rating < lowest.Rating {
			lowest = coffee
		}
	}
	
	stats.AverageRating = float64(totalRating) / float64(len(coffees))
	
	if highest != nil {
		pokemonName := s.getPokemonNameForCoffee(highest.ID, mappings)
		stats.HighestRated = &CoffeeRatingSummary{
			ID:          highest.ID,
			Name:        highest.Name,
			Origin:      highest.Origin,
			Rating:      highest.Rating,
			PokemonName: pokemonName,
		}
	}
	
	if lowest != nil {
		pokemonName := s.getPokemonNameForCoffee(lowest.ID, mappings)
		stats.LowestRated = &CoffeeRatingSummary{
			ID:          lowest.ID,
			Name:        lowest.Name,
			Origin:      lowest.Origin,
			Rating:      lowest.Rating,
			PokemonName: pokemonName,
		}
	}
}

// calculateTypeDistribution calculates Pokemon type distribution
func (s *StatisticsService) calculateTypeDistribution(coffees []models.Coffee, stats *Statistics) {
	for _, coffee := range coffees {
		primaryType, secondaryType, _ := s.mapper.CalculatePokemonTypes(coffee)
		
		stats.TypeDistribution[primaryType]++
		if secondaryType != "" {
			stats.TypeDistribution[secondaryType]++
		}
	}
	
	// Find most common type
	maxCount := 0
	for typeName, count := range stats.TypeDistribution {
		if count > maxCount {
			maxCount = count
			stats.MostCommonType = typeName
		}
	}
}

// calculateOriginStats calculates origin-based statistics
func (s *StatisticsService) calculateOriginStats(coffees []models.Coffee, stats *Statistics) {
	originRatings := make(map[string][]int)
	
	for _, coffee := range coffees {
		if coffee.Origin == "" {
			continue
		}
		stats.OriginDistribution[coffee.Origin]++
		originRatings[coffee.Origin] = append(originRatings[coffee.Origin], coffee.Rating)
	}
	
	// Calculate top origins with average ratings
	type originData struct {
		origin string
		count  int
		avgRating float64
	}
	
	var origins []originData
	for origin, count := range stats.OriginDistribution {
		ratings := originRatings[origin]
		sum := 0
		for _, r := range ratings {
			sum += r
		}
		avg := float64(sum) / float64(len(ratings))
		
		origins = append(origins, originData{
			origin:    origin,
			count:     count,
			avgRating: avg,
		})
	}
	
	// Sort by count (descending)
	sort.Slice(origins, func(i, j int) bool {
		return origins[i].count > origins[j].count
	})
	
	// Take top 5
	limit := 5
	if len(origins) < limit {
		limit = len(origins)
	}
	
	stats.TopOrigins = make([]OriginStat, limit)
	for i := 0; i < limit; i++ {
		stats.TopOrigins[i] = OriginStat{
			Origin:        origins[i].origin,
			Count:         origins[i].count,
			AverageRating: math.Round(origins[i].avgRating*10) / 10,
		}
	}
}

// calculateProcessingStats calculates processing method statistics
func (s *StatisticsService) calculateProcessingStats(coffees []models.Coffee, stats *Statistics) {
	processingRatings := make(map[string][]int)
	processingTypes := make(map[string]map[string]bool)
	
	for _, coffee := range coffees {
		if coffee.ProcessingMethod == "" {
			continue
		}
		
		processingRatings[coffee.ProcessingMethod] = append(
			processingRatings[coffee.ProcessingMethod],
			coffee.Rating,
		)
		
		// Track common types for this processing method
		primaryType, _, _ := s.mapper.CalculatePokemonTypes(coffee)
		if processingTypes[coffee.ProcessingMethod] == nil {
			processingTypes[coffee.ProcessingMethod] = make(map[string]bool)
		}
		processingTypes[coffee.ProcessingMethod][primaryType] = true
	}
	
	for method, ratings := range processingRatings {
		sum := 0
		for _, r := range ratings {
			sum += r
		}
		avg := float64(sum) / float64(len(ratings))
		
		// Get common types (max 3)
		var types []string
		for t := range processingTypes[method] {
			types = append(types, t)
		}
		sort.Strings(types)
		if len(types) > 3 {
			types = types[:3]
		}
		
		stats.ProcessingStats[method] = ProcessingStat{
			Count:         len(ratings),
			AverageRating: math.Round(avg*10) / 10,
			CommonTypes:   types,
		}
	}
}

// calculateRoastDistribution calculates roast level distribution
func (s *StatisticsService) calculateRoastDistribution(coffees []models.Coffee, stats *Statistics) {
	for _, coffee := range coffees {
		if coffee.RoastLevel != "" {
			stats.RoastDistribution[coffee.RoastLevel]++
		}
	}
}

// calculateTraitAverages calculates average tasting traits across all coffees
func (s *StatisticsService) calculateTraitAverages(coffees []models.Coffee, stats *Statistics) {
	if len(coffees) == 0 {
		return
	}
	
	sums := models.TastingTraits{}
	mins := models.TastingTraits{
		BerryIntensity: 10, StonefruitIntensity: 10, RoastIntensity: 10,
		CitrusFruitsIntensity: 10, Bitterness: 10, Florality: 10,
		Spice: 10, Sweetness: 10, AromaticIntensity: 10,
		Savory: 10, Body: 10, Cleanliness: 10,
	}
	maxs := models.TastingTraits{}
	
	for _, coffee := range coffees {
		t := coffee.TastingTraits
		
		sums.BerryIntensity += t.BerryIntensity
		sums.StonefruitIntensity += t.StonefruitIntensity
		sums.RoastIntensity += t.RoastIntensity
		sums.CitrusFruitsIntensity += t.CitrusFruitsIntensity
		sums.Bitterness += t.Bitterness
		sums.Florality += t.Florality
		sums.Spice += t.Spice
		sums.Sweetness += t.Sweetness
		sums.AromaticIntensity += t.AromaticIntensity
		sums.Savory += t.Savory
		sums.Body += t.Body
		sums.Cleanliness += t.Cleanliness
		
		// Track min/max
		mins.BerryIntensity = minInt(mins.BerryIntensity, t.BerryIntensity)
		maxs.BerryIntensity = maxInt(maxs.BerryIntensity, t.BerryIntensity)
		mins.StonefruitIntensity = minInt(mins.StonefruitIntensity, t.StonefruitIntensity)
		maxs.StonefruitIntensity = maxInt(maxs.StonefruitIntensity, t.StonefruitIntensity)
		mins.RoastIntensity = minInt(mins.RoastIntensity, t.RoastIntensity)
		maxs.RoastIntensity = maxInt(maxs.RoastIntensity, t.RoastIntensity)
		mins.CitrusFruitsIntensity = minInt(mins.CitrusFruitsIntensity, t.CitrusFruitsIntensity)
		maxs.CitrusFruitsIntensity = maxInt(maxs.CitrusFruitsIntensity, t.CitrusFruitsIntensity)
		mins.Bitterness = minInt(mins.Bitterness, t.Bitterness)
		maxs.Bitterness = maxInt(maxs.Bitterness, t.Bitterness)
		mins.Florality = minInt(mins.Florality, t.Florality)
		maxs.Florality = maxInt(maxs.Florality, t.Florality)
		mins.Spice = minInt(mins.Spice, t.Spice)
		maxs.Spice = maxInt(maxs.Spice, t.Spice)
		mins.Sweetness = minInt(mins.Sweetness, t.Sweetness)
		maxs.Sweetness = maxInt(maxs.Sweetness, t.Sweetness)
		mins.AromaticIntensity = minInt(mins.AromaticIntensity, t.AromaticIntensity)
		maxs.AromaticIntensity = maxInt(maxs.AromaticIntensity, t.AromaticIntensity)
		mins.Savory = minInt(mins.Savory, t.Savory)
		maxs.Savory = maxInt(maxs.Savory, t.Savory)
		mins.Body = minInt(mins.Body, t.Body)
		maxs.Body = maxInt(maxs.Body, t.Body)
		mins.Cleanliness = minInt(mins.Cleanliness, t.Cleanliness)
		maxs.Cleanliness = maxInt(maxs.Cleanliness, t.Cleanliness)
	}
	
	count := len(coffees)
	stats.TraitAverages = models.TastingTraits{
		BerryIntensity:        sums.BerryIntensity / count,
		StonefruitIntensity:   sums.StonefruitIntensity / count,
		RoastIntensity:        sums.RoastIntensity / count,
		CitrusFruitsIntensity: sums.CitrusFruitsIntensity / count,
		Bitterness:            sums.Bitterness / count,
		Florality:             sums.Florality / count,
		Spice:                 sums.Spice / count,
		Sweetness:             sums.Sweetness / count,
		AromaticIntensity:     sums.AromaticIntensity / count,
		Savory:                sums.Savory / count,
		Body:                  sums.Body / count,
		Cleanliness:           sums.Cleanliness / count,
	}
	
	stats.TraitRanges = TraitRanges{
		BerryRange:      Range{Min: mins.BerryIntensity, Max: maxs.BerryIntensity},
		StonefruitRange: Range{Min: mins.StonefruitIntensity, Max: maxs.StonefruitIntensity},
		RoastRange:      Range{Min: mins.RoastIntensity, Max: maxs.RoastIntensity},
		CitrusRange:     Range{Min: mins.CitrusFruitsIntensity, Max: maxs.CitrusFruitsIntensity},
		BitternessRange: Range{Min: mins.Bitterness, Max: maxs.Bitterness},
		FloralityRange:  Range{Min: mins.Florality, Max: maxs.Florality},
		SpiceRange:      Range{Min: mins.Spice, Max: maxs.Spice},
		SweetnessRange:  Range{Min: mins.Sweetness, Max: maxs.Sweetness},
		AromaticRange:   Range{Min: mins.AromaticIntensity, Max: maxs.AromaticIntensity},
		SavoryRange:     Range{Min: mins.Savory, Max: maxs.Savory},
		BodyRange:       Range{Min: mins.Body, Max: maxs.Body},
		CleanlinessRange: Range{Min: mins.Cleanliness, Max: maxs.Cleanliness},
	}
}

// calculateBrewerStats calculates brewer/dripper statistics
func (s *StatisticsService) calculateBrewerStats(coffees []models.Coffee, stats *Statistics) {
	brewerRatings := make(map[string][]int)
	brewerTimes := make(map[string][]float64)
	
	for _, coffee := range coffees {
		if coffee.Dripper == "" {
			continue
		}
		
		brewerRatings[coffee.Dripper] = append(brewerRatings[coffee.Dripper], coffee.Rating)
		
		// Calculate brew time in seconds
		brewTime := float64(coffee.EndTime.Minutes*60 + coffee.EndTime.Seconds)
		if brewTime > 0 {
			brewerTimes[coffee.Dripper] = append(brewerTimes[coffee.Dripper], brewTime)
		}
	}
	
	for brewer, ratings := range brewerRatings {
		sum := 0
		for _, r := range ratings {
			sum += r
		}
		avg := float64(sum) / float64(len(ratings))
		
		// Calculate average brew time
		avgTime := 0.0
		if times, ok := brewerTimes[brewer]; ok && len(times) > 0 {
			timeSum := 0.0
			for _, t := range times {
				timeSum += t
			}
			avgTime = timeSum / float64(len(times))
		}
		
		stats.BrewerStats[brewer] = BrewerStat{
			Count:         len(ratings),
			AverageRating: math.Round(avg*10) / 10,
			AvgBrewTime:   math.Round(avgTime*10) / 10,
		}
	}
}

// calculateConfidenceMetrics calculates Pokemon mapping confidence metrics
func (s *StatisticsService) calculateConfidenceMetrics(mappings []models.CoffeePokemon, stats *Statistics) {
	if len(mappings) == 0 {
		return
	}
	
	totalConfidence := 0.0
	highConfidence := 0
	
	for _, mapping := range mappings {
		totalConfidence += mapping.MappingConfidence
		if mapping.MappingConfidence >= 0.8 {
			highConfidence++
		}
	}
	
	stats.AverageConfidence = math.Round((totalConfidence/float64(len(mappings)))*100) / 100
	stats.HighConfidencePairings = highConfidence
}

// getPokemonNameForCoffee helper to get Pokemon name for a coffee ID
func (s *StatisticsService) getPokemonNameForCoffee(coffeeID string, mappings []models.CoffeePokemon) string {
	for _, mapping := range mappings {
		if mapping.CoffeeID == coffeeID {
			return mapping.PokemonName
		}
	}
	return ""
}

// minInt returns minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}