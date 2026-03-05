package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-coffee-log/models"
	"log"
	"os/exec"
	"strings"
	"time"
)

// ClaudeService handles Pokemon selection via the Claude CLI
type ClaudeService struct {
	model   string
	timeout time.Duration
}

// NewClaudeService creates a new Claude CLI service
func NewClaudeService(model string) *ClaudeService {
	return &ClaudeService{
		model:   model,
		timeout: 120 * time.Second,
	}
}

// ClaudeResponse is the JSON response from Claude for Pokemon selection
type ClaudeResponse struct {
	PokemonName string  `json:"pokemon_name"`
	PokemonID   int     `json:"pokemon_id"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// SelectPokemon asks Claude to pick a Pokemon for a coffee from available Pokemon
func (s *ClaudeService) SelectPokemon(
	coffee models.Coffee,
	traits models.TastingTraits,
	combinedNotes []string,
	typeHints map[string]float64,
	availablePokemon []models.Pokemon,
) (*ClaudeResponse, error) {
	prompt := s.buildPrompt(coffee, traits, combinedNotes, typeHints, availablePokemon)

	log.Printf("Calling Claude CLI for Pokemon selection (coffee: %s, %d available Pokemon)", coffee.Name, len(availablePokemon))

	// Build the command: pipe prompt via stdin to handle long prompts
	cmd := exec.Command("claude", "-p", "--output-format", "json", "--model", s.model, "--no-session-persistence")
	cmd.Stdin = bytes.NewBufferString(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("claude CLI failed: %w (stderr: %s)", err, stderr.String())
		}
	case <-time.After(s.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("claude CLI timed out after %s", s.timeout)
	}

	// Parse the JSON output from Claude CLI
	// The --output-format json wraps the response; extract the result text
	output := stdout.Bytes()

	var cliResponse struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &cliResponse); err != nil {
		// Maybe it returned the JSON directly
		return s.parseResponse(output, availablePokemon)
	}

	if cliResponse.Result != "" {
		return s.parseResponse([]byte(cliResponse.Result), availablePokemon)
	}

	// Try parsing the raw output
	return s.parseResponse(output, availablePokemon)
}

// parseResponse extracts the ClaudeResponse from Claude's output
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
		return nil, fmt.Errorf("failed to parse Claude response as JSON: %w\nraw: %s", err, string(data))
	}

	// Validate that the selected Pokemon is actually in the available list
	valid := false
	for _, p := range availablePokemon {
		if p.ID == resp.PokemonID {
			valid = true
			// Ensure the name matches the ID
			resp.PokemonName = p.Name
			break
		}
	}
	if !valid {
		// Try matching by name as fallback
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
		return nil, fmt.Errorf("Claude selected Pokemon %q (ID %d) which is not in the available list", resp.PokemonName, resp.PokemonID)
	}

	return &resp, nil
}

// buildPrompt creates the prompt for Claude to select a Pokemon
func (s *ClaudeService) buildPrompt(
	coffee models.Coffee,
	traits models.TastingTraits,
	combinedNotes []string,
	typeHints map[string]float64,
	availablePokemon []models.Pokemon,
) string {
	// Build type hints string (sorted by score)
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

	// Build available Pokemon list
	var pokemonList []string
	for _, p := range availablePokemon {
		pokemonList = append(pokemonList, fmt.Sprintf("- #%d %s (%s)", p.ID, p.Name, p.Type))
	}

	// Build notes string
	notesStr := "none recorded"
	filteredNotes := make([]string, 0)
	for _, n := range combinedNotes {
		if n != "" {
			filteredNotes = append(filteredNotes, n)
		}
	}
	if len(filteredNotes) > 0 {
		notesStr = strings.Join(filteredNotes, ", ")
	}

	prompt := fmt.Sprintf(`You are the CoffeeDex - a fun Pokemon-themed coffee journal. A user just logged enough brews of a coffee to unlock a Pokemon companion for it. Your job is to pick a Pokemon and write a fun, entertaining Pokedex-style entry.

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

Respond with ONLY this JSON:
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

	return prompt
}
