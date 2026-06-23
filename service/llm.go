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

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
const defaultModel = "anthropic/claude-sonnet-4-5"

// ClaudeService handles Pokemon selection via OpenRouter
type ClaudeService struct {
	apiKey  string
	model   string
	timeout time.Duration
	client  *http.Client
}

// NewClaudeService creates a new OpenRouter-backed service
func NewClaudeService(apiKey string) *ClaudeService {
	return &ClaudeService{
		apiKey:  apiKey,
		model:   defaultModel,
		timeout: 60 * time.Second,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// ClaudeResponse is the structured response for Pokemon selection
type ClaudeResponse struct {
	PokemonName string  `json:"pokemon_name"`
	PokemonID   int     `json:"pokemon_id"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// SelectPokemon asks the LLM to pick a Pokemon for a coffee
func (s *ClaudeService) SelectPokemon(
	coffee models.Coffee,
	traits models.TastingTraits,
	combinedNotes []string,
	typeHints map[string]float64,
	availablePokemon []models.Pokemon,
) (*ClaudeResponse, error) {
	prompt := s.buildPrompt(coffee, traits, combinedNotes, typeHints, availablePokemon)

	log.Printf("Calling OpenRouter (%s) for Pokemon selection (coffee: %s, %d available Pokemon)", s.model, coffee.Name, len(availablePokemon))

	reqBody, err := json.Marshal(map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.8,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/coffee-dex")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenRouter returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	content := apiResp.Choices[0].Message.Content
	return s.parseResponse([]byte(content), availablePokemon)
}

// parseResponse extracts the ClaudeResponse from the LLM output
func (s *ClaudeService) parseResponse(data []byte, availablePokemon []models.Pokemon) (*ClaudeResponse, error) {
	text := string(data)

	// Strip markdown code blocks if present
	text = strings.ReplaceAll(text, "```json", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.TrimSpace(text)

	// Extract JSON object
	if start := strings.Index(text, "{"); start >= 0 {
		if end := strings.LastIndex(text, "}"); end > start {
			text = text[start : end+1]
		}
	}

	var resp ClaudeResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w\nraw: %s", err, string(data))
	}

	// Validate that the selected Pokemon is in the available list
	valid := false
	for _, p := range availablePokemon {
		if p.ID == resp.PokemonID {
			valid = true
			resp.PokemonName = p.Name
			break
		}
	}
	if !valid {
		for _, p := range availablePokemon {
			if strings.EqualFold(p.Name, resp.PokemonName) {
				resp.PokemonID = p.ID
				resp.PokemonName = p.Name
				valid = true
				break
			}
		}
	}
	if !valid {
		return nil, fmt.Errorf("LLM selected Pokemon %q (ID %d) which is not in the available list", resp.PokemonName, resp.PokemonID)
	}

	return &resp, nil
}

// buildPrompt creates the prompt for Pokemon selection
func (s *ClaudeService) buildPrompt(
	coffee models.Coffee,
	traits models.TastingTraits,
	combinedNotes []string,
	typeHints map[string]float64,
	availablePokemon []models.Pokemon,
) string {
	var typeHintParts []string
	for typeName, score := range typeHints {
		if score > 0.3 {
			typeHintParts = append(typeHintParts, fmt.Sprintf("%s (%.0f%%)", typeName, score*100))
		}
	}
	typeHintStr := "None strong enough"
	if len(typeHintParts) > 0 {
		typeHintStr = strings.Join(typeHintParts, ", ")
	}

	var pokemonList []string
	for _, p := range availablePokemon {
		pokemonList = append(pokemonList, fmt.Sprintf("- #%d %s (%s)", p.ID, p.Name, p.Type))
	}

	filteredNotes := make([]string, 0)
	for _, n := range combinedNotes {
		if n != "" {
			filteredNotes = append(filteredNotes, n)
		}
	}
	notesStr := "none recorded"
	if len(filteredNotes) > 0 {
		notesStr = strings.Join(filteredNotes, ", ")
	}

	return fmt.Sprintf(`You are the CoffeeDex - a fun Pokemon-themed coffee journal. A user just logged enough brews of a coffee to unlock a Pokemon companion for it. Your job is to pick a Pokemon and write a fun, entertaining Pokedex-style entry.

COFFEE:
- Name: %s
- Origin: %s
- Roaster: %s
- Variety: %s
- Roast Level: %s
- Processing: %s

TASTING NOTES FROM BREWS: %s

FLAVOR SCORES (0-10, -1 means not scored):
Sweetness: %d, Bitterness: %d, Citrus: %d, Berry: %d, Stonefruit: %d, Florality: %d, Roast: %d, Spice: %d, Aroma: %d, Savory: %d, Body: %d, Cleanliness: %d

TYPE SUGGESTIONS (hints from flavor mapping, not constraints): %s

AVAILABLE POKEMON (pick ONLY from this list):
%s

YOUR TASK:
1. Pick the Pokemon that best vibes with this coffee's flavor profile and personality.
2. Write a fun 2-4 sentence Pokedex-style description that:
   - References specific tasting notes the user logged
   - Connects them to Pokemon lore (anime moments, game abilities, Pokedex entries, type matchups, etc.)
   - Is entertaining and playful, not dry or technical
   - Makes the user smile and feel like the Pokemon "gets" their coffee

The type suggestions are just hints - feel free to pick any Pokemon from the available list that feels right.

Respond with ONLY this JSON (no markdown, no explanation):
{
  "pokemon_name": "ExactName",
  "pokemon_id": 123,
  "confidence": 0.85,
  "description": "Your fun description here"
}`,
		coffee.Name, coffee.Origin, coffee.Roaster, coffee.Variety,
		coffee.RoastLevel, coffee.ProcessingMethod,
		notesStr,
		traits.Sweetness, traits.Bitterness, traits.CitrusFruitsIntensity, traits.BerryIntensity,
		traits.StonefruitIntensity, traits.Florality, traits.RoastIntensity, traits.Spice,
		traits.AromaticIntensity, traits.Savory, traits.Body, traits.Cleanliness,
		typeHintStr,
		strings.Join(pokemonList, "\n"))
}
