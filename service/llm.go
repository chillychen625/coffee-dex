package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// LLMService handles communication with Ollama for Pokemon mapping
type LLMService struct {
	client  *http.Client
	baseURL string
	model   string
	timeout time.Duration
}

// NewLLMService creates a new LLM service for Ollama
func NewLLMService(baseURL string, model string) *LLMService {
	return &LLMService{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
		model:   model,
		timeout: 30 * time.Second,
	}
}

// MapCoffeeToPokemon maps coffee to Pokemon using LLM
func (s *LLMService) MapCoffeeToPokemon(coffee models.Coffee, candidates []models.Pokemon) (*models.LLMMappingResponse, error) {
	prompt := s.buildPrompt(coffee, candidates)
	
	payload := map[string]interface{}{
		"model":  s.model,
		"prompt": prompt,
		"stream": false,
		"format": "json",
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", s.baseURL+"/api/generate", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: s.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Response string `json:"response"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode LLM response: %w", err)
	}
	
	// Parse the JSON response from LLM
	return s.parseLLMResponse(response.Response)
}

// buildPrompt creates the prompt for LLM mapping
func (s *LLMService) buildPrompt(coffee models.Coffee, candidates []models.Pokemon) string {
	var candidateNames []string
	for _, candidate := range candidates {
		candidateNames = append(candidateNames, candidate.Name)
	}
	
	traitDescription := s.formatTraits(coffee.TastingTraits)
	
	prompt := fmt.Sprintf(`You are a Pokemon expert specializing in coffee-Pokemon mappings. 
Given a coffee's characteristics, select the best Gen 1 Pokemon match and write a Pokedex-style description.

Coffee: %s from %s
Tasting Notes: %s
Dominant Traits: %s

Available Pokemon: %s

Respond with ONLY valid JSON:
{
  "selected_pokemon": "exact_pokemon_name",
  "confidence": 0.95,
  "description": "Pokedex-style description connecting coffee traits to Pokemon characteristics",
  "trait_mapping": [
    {"trait": "sweetness", "pokemon_stat": "HP", "reasoning": "sweet coffee provides sustained energy"},
    {"trait": "bitterness", "pokemon_stat": "Attack", "reasoning": "bitterness represents bold, attacking flavors"}
  ]
}`, coffee.Name, coffee.Origin, strings.Join(coffee.TastingNotes[:], ", "), traitDescription, strings.Join(candidateNames, ", "))
	
	return prompt
}

// formatTraits formats coffee traits for LLM prompt
func (s *LLMService) formatTraits(traits models.TastingTraits) string {
	highTraits := []string{}
	
	if traits.Sweetness >= 7 {
		highTraits = append(highTraits, fmt.Sprintf("high sweetness (%d)", traits.Sweetness))
	}
	if traits.Bitterness >= 7 {
		highTraits = append(highTraits, fmt.Sprintf("high bitterness (%d)", traits.Bitterness))
	}
	if traits.CitrusFruitsIntensity >= 7 {
		highTraits = append(highTraits, fmt.Sprintf("high citrus (%d)", traits.CitrusFruitsIntensity))
	}
	if traits.Florality >= 7 {
		highTraits = append(highTraits, fmt.Sprintf("high florality (%d)", traits.Florality))
	}
	if traits.Body >= 7 {
		highTraits = append(highTraits, fmt.Sprintf("full body (%d)", traits.Body))
	}
	if traits.AromaticIntensity >= 7 {
		highTraits = append(highTraits, fmt.Sprintf("high aroma (%d)", traits.AromaticIntensity))
	}
	
	if len(highTraits) == 0 {
		return "balanced traits"
	}
	
	return strings.Join(highTraits, ", ")
}

// parseLLMResponse parses the LLM response
func (s *LLMService) parseLLMResponse(response string) (*models.LLMMappingResponse, error) {
	// Clean up the response to extract JSON
	response = strings.TrimSpace(response)
	
	// Remove any markdown code blocks
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	
	var mappingResponse models.LLMMappingResponse
	if err := json.Unmarshal([]byte(response), &mappingResponse); err != nil {
		// Try to fix common JSON issues
		log.Printf("Failed to parse LLM response as JSON: %s", response)
		
		// Fallback: try to extract Pokemon name using regex-like parsing
		return s.fallbackParse(response), nil
	}
	
	return &mappingResponse, nil
}

// fallbackParse provides a basic fallback when JSON parsing fails
func (s *LLMService) fallbackParse(response string) *models.LLMMappingResponse {
	// Simple fallback - look for common Pokemon names
	pokemonNames := []string{"bulbasaur", "charmander", "squirtle", "pikachu", "jigglypuff"}
	
	var selectedPokemon string
	for _, name := range pokemonNames {
		if strings.Contains(strings.ToLower(response), name) {
			selectedPokemon = name
			break
		}
	}
	
	if selectedPokemon == "" {
		selectedPokemon = "bulbasaur" // Default fallback
	}
	
	return &models.LLMMappingResponse{
		SelectedPokemon: selectedPokemon,
		Confidence:      0.5,
		Description:     "Fallback mapping due to parsing error",
		TraitMapping: []models.TraitMapping{
			{Trait: "general", PokemonStat: "HP", Reasoning: "Basic fallback mapping"},
		},
	}
}

// TestConnection tests the connection to LLM service
func (s *LLMService) TestConnection() error {
	req, err := http.NewRequest("GET", s.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to LLM: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LLM service returned status %d", resp.StatusCode)
	}
	
	return nil
}