package service

import (
	"fmt"
	"go-coffee-log/models"
	"go-coffee-log/storage"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PokemonService handles business logic for Pokemon operations
type PokemonService struct {
	storage      storage.PokemonStorage
	coffeeService *CoffeeService
	llmService   *LLMService
}

// NewPokemonService creates a new Pokemon service
func NewPokemonService(
	pokemonStorage storage.PokemonStorage,
	coffeeService *CoffeeService,
	llmService *LLMService,
) *PokemonService {
	return &PokemonService{
		storage:      pokemonStorage,
		coffeeService: coffeeService,
		llmService:   llmService,
	}
}

// MapCoffeeToPokemon maps a coffee to a Pokemon using rules and LLM
func (s *PokemonService) MapCoffeeToPokemon(coffee models.Coffee) (*models.CoffeePokemon, error) {
	// 1. Get rule-based candidates
	candidates := s.getRuleBasedCandidates(coffee)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no Pokemon candidates found for coffee characteristics")
	}

	// 2. Try LLM mapping first
	var selectedPokemon *models.Pokemon
	var confidence float64
	var description string
	var traitMapping []models.TraitMapping

	if s.llmService != nil {
		llmResponse, err := s.llmService.MapCoffeeToPokemon(coffee, candidates)
		if err != nil {
			log.Printf("LLM mapping failed, using rules: %v", err)
			selectedPokemon, confidence, description, traitMapping = s.getRuleBasedMapping(coffee)
		} else {
			// Find the Pokemon by name from LLM response
			for _, candidate := range candidates {
				if strings.ToLower(candidate.Name) == strings.ToLower(llmResponse.SelectedPokemon) {
					selectedPokemon = &candidate
					break
				}
			}
			if selectedPokemon == nil {
				log.Printf("LLM selected unknown Pokemon: %s, using rules", llmResponse.SelectedPokemon)
				selectedPokemon, confidence, description, traitMapping = s.getRuleBasedMapping(coffee)
			} else {
				confidence = llmResponse.Confidence
				description = llmResponse.Description
				traitMapping = llmResponse.TraitMapping
			}
		}
	} else {
		selectedPokemon, confidence, description, traitMapping = s.getRuleBasedMapping(coffee)
	}

	// 3. Ensure uniqueness
	finalPokemon, err := s.ensureUniquePokemon(coffee.ID, *selectedPokemon)
	if err != nil {
		return nil, fmt.Errorf("no unique Pokemon available: %w", err)
	}

	// 4. Create mapping
	mapping := &models.CoffeePokemon{
		ID:                uuid.New().String(),
		CoffeeID:          coffee.ID,
		PokemonID:         finalPokemon.ID,
		PokemonName:       finalPokemon.Name,
		Nickname:          "",
		Level:             s.calculateLevel(coffee.Rating),
		MappingConfidence: confidence,
		LLMDescription:    description,
		TraitMapping:      traitMapping,
		CreatedAt:         time.Now(),
	}

	if err := s.storage.CreateCoffeePokemon(*mapping); err != nil {
		return nil, fmt.Errorf("failed to create Pokemon mapping: %w", err)
	}
	return mapping, nil
}

// getRuleBasedCandidates gets Pokemon candidates based on coffee traits
func (s *PokemonService) getRuleBasedCandidates(coffee models.Coffee) []models.Pokemon {
	pokemonType := s.determinePrimaryType(coffee.TastingTraits)
	
	// Get Pokemon of the determined type
	pokemons, err := s.storage.GetPokemonByType(pokemonType)
	if err != nil {
		log.Printf("Failed to get Pokemon by type %s: %v", pokemonType, err)
		return []models.Pokemon{}
	}

	// Filter by generation based on coffee complexity
	_ = s.determineGeneration(coffee) // Currently using Gen 1 only
	filtered := make([]models.Pokemon, 0)
	for _, pokemon := range pokemons {
		// For now, all Pokemon in storage are Gen 1, so we'll adjust stats based on complexity
		if s.isCoffeeComplexityMatch(pokemon, coffee) {
			filtered = append(filtered, pokemon)
		}
	}

	// If no specific matches, return some general Pokemon
	if len(filtered) == 0 {
		return pokemons[:min(5, len(pokemons))]
	}

	return filtered[:min(8, len(filtered))]
}

// determinePrimaryType determines Pokemon type based on coffee traits
func (s *PokemonService) determinePrimaryType(traits models.TastingTraits) string {
	// Analyze dominant traits
	if traits.Sweetness >= 7 && traits.Bitterness <= 3 {
		return "Normal" // Sweet, pleasant nature
	}
	if traits.Bitterness >= 7 && traits.RoastIntensity >= 6 {
		return "Fire" // Bold, intense characteristics
	}
	if traits.CitrusFruitsIntensity >= 7 && traits.Florality >= 6 {
		return "Grass" // Bright, aromatic, natural
	}
	if traits.BerryIntensity >= 6 && traits.StonefruitIntensity >= 5 {
		return "Poison" // Fruity, complex flavors
	}
	if traits.Spice >= 6 && traits.Body >= 6 {
		return "Ground" // Earthy, warming characteristics
	}
	if traits.Savory >= 6 && traits.RoastIntensity >= 5 {
		return "Electric" // Bold, structured, intense
	}
	if traits.Cleanliness >= 7 && traits.Bitterness <= 4 {
		return "Water" // Clean, pure, refreshing
	}
	if traits.AromaticIntensity >= 7 && traits.Florality >= 6 {
		return "Psychic" // Complex, mind-blowing aromas
	}

	// Default fallback
	return "Normal"
}

// determineGeneration determines Pokemon generation based on coffee complexity
func (s *PokemonService) determineGeneration(coffee models.Coffee) int {
	variance := s.calculateTraitVariance(coffee.TastingTraits)
	
	if variance < 15 {
		return 1 // Common Pokemon - basic coffees
	} else if variance < 30 {
		return 1 // Uncommon Pokemon - good coffees
	} else if variance < 45 {
		return 1 // Rare Pokemon - excellent coffees
	} else {
		return 1 // Legendary Pokemon - exceptional coffees
	}
}

// isCoffeeComplexityMatch checks if Pokemon stats match coffee complexity
func (s *PokemonService) isCoffeeComplexityMatch(pokemon models.Pokemon, coffee models.Coffee) bool {
	variance := s.calculateTraitVariance(coffee.TastingTraits)
	avgStats := (pokemon.BaseStats.HP + pokemon.BaseStats.Attack + pokemon.BaseStats.Defense + 
		pokemon.BaseStats.Speed + pokemon.BaseStats.Special) / 5
	
	// Match complexity to Pokemon stats
	if variance < 20 && avgStats < 50 {
		return true // Basic Pokemon for simple coffee
	} else if variance < 35 && avgStats < 70 {
		return true // Medium Pokemon for medium coffee
	} else if variance >= 35 {
		return true // Strong Pokemon for complex coffee
	}
	
	return false
}

// getRuleBasedMapping gets Pokemon mapping using only rules
func (s *PokemonService) getRuleBasedMapping(coffee models.Coffee) (*models.Pokemon, float64, string, []models.TraitMapping) {
	candidates := s.getRuleBasedCandidates(coffee)
	if len(candidates) == 0 {
		// Fallback to a basic Pokemon
		return &models.Pokemon{
			ID:          1, // Bulbasaur
			Name:        "Bulbasaur",
			Type:        "Grass/Poison",
			Description: "A basic Pokemon for coffee mapping",
		}, 0.5, "Basic rule-based mapping", []models.TraitMapping{}
	}

	// Select the first candidate for now
	selected := candidates[0]
	confidence := 0.7
	description := fmt.Sprintf("Rule-based mapping: %s with type %s matches coffee characteristics", selected.Name, selected.Type)
	
	traitMapping := []models.TraitMapping{
		{Trait: "sweetness", PokemonStat: "HP", Reasoning: "Sweet coffee provides sustained energy"},
		{Trait: "bitterness", PokemonStat: "Attack", Reasoning: "Bitterness represents bold, attacking flavors"},
	}

	return &selected, confidence, description, traitMapping
}

// ensureUniquePokemon ensures each Pokemon is unique
func (s *PokemonService) ensureUniquePokemon(coffeeID string, pokemon models.Pokemon) (*models.Pokemon, error) {
	used, err := s.storage.IsPokemonUsed(pokemon.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check Pokemon usage: %w", err)
	}

	if !used {
		return &pokemon, nil
	}

	// Find alternative Pokemon with similar characteristics
	alternatives, err := s.storage.GetPokemonByType(pokemon.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get alternative Pokemon: %w", err)
	}

	for _, alt := range alternatives {
		altUsed, err := s.storage.IsPokemonUsed(alt.ID)
		if err != nil {
			continue
		}
		if !altUsed {
			return &alt, nil
		}
	}

	// If no alternatives, return original (will fail on database constraint)
	return &pokemon, fmt.Errorf("Pokemon %s already used and no alternatives available", pokemon.Name)
}

// calculateLevel calculates Pokemon level based on coffee rating
func (s *PokemonService) calculateLevel(rating int) int {
	// Level 1-50 based on rating 0-10
	return rating * 5
}

// calculateTraitVariance calculates variance in coffee traits
func (s *PokemonService) calculateTraitVariance(traits models.TastingTraits) int {
	traitValues := []int{
		traits.BerryIntensity, traits.StonefruitIntensity, traits.RoastIntensity,
		traits.CitrusFruitsIntensity, traits.Bitterness, traits.Florality,
		traits.Spice, traits.Sweetness, traits.AromaticIntensity,
		traits.Savory, traits.Body, traits.Cleanliness,
	}

	// Calculate mean
	sum := 0
	for _, val := range traitValues {
		sum += val
	}
	mean := sum / len(traitValues)

	// Calculate variance
	variance := 0
	for _, val := range traitValues {
		diff := val - mean
		variance += diff * diff
	}

	return variance / len(traitValues)
}

// GetCoffeePokemon gets Pokemon mapping for a specific coffee
func (s *PokemonService) GetCoffeePokemon(coffeeID string) (*models.CoffeePokemon, error) {
	return s.storage.GetCoffeePokemon(coffeeID)
}

// GetAllCoffeePokemon gets all coffee-Pokemon mappings
func (s *PokemonService) GetAllCoffeePokemon() ([]models.CoffeePokemon, error) {
	return s.storage.GetAllCoffeePokemon()
}

// UpdateNickname updates Pokemon nickname
func (s *PokemonService) UpdateNickname(coffeeID, nickname string) error {
	return s.storage.UpdateCoffeePokemonNickname(coffeeID, nickname)
}

// InitializePokemonData initializes Pokemon data in the database
func (s *PokemonService) InitializePokemonData() error {
	// Check if Pokemon data already exists
	existing, err := s.storage.GetAllPokemon()
	if err == nil && len(existing) > 0 {
		log.Println("Pokemon data already exists, skipping initialization")
		return nil
	}

	// Seed Gen 1 Pokemon data
	pokemonData := s.getGen1PokemonData()
	
	// Note: This would typically be done through a separate seeding script
	// For now, just log that initialization is needed
	log.Printf("Need to seed %d Gen 1 Pokemon", len(pokemonData))
	
	return nil
}

// getGen1PokemonData returns Gen 1 Pokemon data (simplified for demo)
func (s *PokemonService) getGen1PokemonData() []models.Pokemon {
	return []models.Pokemon{
		{
			ID:          1,
			Name:        "Bulbasaur",
			Type:        "Grass/Poison",
			SpritePath:  "/sprites/001-bulbasaur.png",
			Description: "A strange seed was planted on its back at birth. The plant sprouts and grows with this Pokemon.",
			BaseStats: models.Stats{HP: 45, Attack: 49, Defense: 49, Speed: 45, Special: 65},
		},
		{
			ID:          4,
			Name:        "Charmander",
			Type:        "Fire",
			SpritePath:  "/sprites/004-charmander.png",
			Description: "Obviously prefers hot places. When it rains, steam is said to spout from the tip of its tail.",
			BaseStats: models.Stats{HP: 39, Attack: 52, Defense: 43, Speed: 65, Special: 50},
		},
		{
			ID:          7,
			Name:        "Squirtle",
			Type:        "Water",
			SpritePath:  "/sprites/007-squirtle.png",
			Description: "After birth, its back swells and hardens into a shell. Powerfully sprays foam from its mouth.",
			BaseStats: models.Stats{HP: 44, Attack: 48, Defense: 65, Speed: 43, Special: 50},
		},
		{
			ID:          25,
			Name:        "Pikachu",
			Type:        "Electric",
			SpritePath:  "/sprites/025-pikachu.png",
			Description: "When several of these Pokemon gather, their electricity could build and cause lightning storms.",
			BaseStats: models.Stats{HP: 35, Attack: 55, Defense: 40, Speed: 90, Special: 50},
		},
		{
			ID:          39,
			Name:        "Jigglypuff",
			Type:        "Normal/Fairy",
			SpritePath:  "/sprites/039-jigglypuff.png",
			Description: "When its huge eyes light up, it sings a mysteriously soothing melody that lulls its enemies to sleep.",
			BaseStats: models.Stats{HP: 115, Attack: 45, Defense: 20, Speed: 20, Special: 25},
		},
		// Add more Pokemon as needed
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}