// fix-descriptions patches weak Pokemon descriptions in the database by generating
// new ones via OpenRouter using each coffee's actual brew notes and traits.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
const model = "anthropic/claude-sonnet-4-5"

type CoffeeRow struct {
	CoffeePokemonID string
	CoffeeName      string
	Origin          string
	Roaster         string
	Variety         string
	RoastLevel      string
	Processing      string
	PokemonName     string
	PokemonType     string
	OldDescription  string
	Notes           []string
	Traits          []map[string]interface{}
}

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY env var required")
	}

	dbPath := "./coffee-dex.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT cp.id, c.name, c.origin, c.roaster, c.variety, c.roast_level, c.processing_method,
		       p.name, p.type, cp.llm_description
		FROM coffee_pokemon cp
		JOIN coffees c ON cp.coffee_id = c.id
		JOIN pokemons p ON cp.pokemon_id = p.id
		WHERE cp.llm_description LIKE '%Fallback mapping%'
		   OR cp.llm_description LIKE '%Type-based mapping%'
		   OR cp.llm_description LIKE '%unique flavor profile%'
		   OR cp.llm_description LIKE '%unique Pokemon mapped%'
		   OR cp.llm_description LIKE '%A Normal-type Pokemon%'
		   OR cp.llm_description LIKE '%A unique Pokemon%'
		   OR cp.llm_description LIKE '%was matched to this coffee%'
		ORDER BY c.name`)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var coffees []CoffeeRow
	for rows.Next() {
		var r CoffeeRow
		if err := rows.Scan(&r.CoffeePokemonID, &r.CoffeeName, &r.Origin, &r.Roaster,
			&r.Variety, &r.RoastLevel, &r.Processing,
			&r.PokemonName, &r.PokemonType, &r.OldDescription); err != nil {
			log.Printf("scan: %v", err)
			continue
		}
		coffees = append(coffees, r)
	}
	rows.Close()

	// Load brew data for each coffee
	for i := range coffees {
		brewRows, err := db.Query(`
			SELECT tasting_notes, tasting_traits
			FROM brews
			WHERE coffee_id = (SELECT id FROM coffees WHERE name = ?)
			  AND (tasting_notes IS NOT NULL OR tasting_traits IS NOT NULL)`,
			coffees[i].CoffeeName)
		if err != nil {
			continue
		}
		for brewRows.Next() {
			var notes sql.NullString
			var traits sql.NullString
			if err := brewRows.Scan(&notes, &traits); err != nil {
				continue
			}
			if notes.Valid && notes.String != "" {
				coffees[i].Notes = append(coffees[i].Notes, notes.String)
			}
			if traits.Valid && traits.String != "" {
				var t map[string]interface{}
				if err := json.Unmarshal([]byte(traits.String), &t); err == nil {
					coffees[i].Traits = append(coffees[i].Traits, t)
				}
			}
		}
		brewRows.Close()
	}

	client := &http.Client{Timeout: 60 * time.Second}
	fmt.Printf("Fixing %d weak descriptions...\n\n", len(coffees))

	for i, c := range coffees {
		fmt.Printf("[%d/%d] %s → %s (%s)\n", i+1, len(coffees), c.CoffeeName, c.PokemonName, c.PokemonType)

		desc, err := generateDescription(client, apiKey, c)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}

		_, err = db.Exec(`UPDATE coffee_pokemon SET llm_description = ? WHERE id = ?`, desc, c.CoffeePokemonID)
		if err != nil {
			fmt.Printf("  DB ERROR: %v\n", err)
			continue
		}

		fmt.Printf("  → %s\n\n", desc)
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("Done!")
}

func generateDescription(client *http.Client, apiKey string, c CoffeeRow) (string, error) {
	notesStr := "none recorded"
	if len(c.Notes) > 0 {
		notesStr = strings.Join(c.Notes, " | ")
	}

	traitsStr := "none"
	if len(c.Traits) > 0 {
		// Average traits
		avg := make(map[string]float64)
		count := make(map[string]int)
		for _, t := range c.Traits {
			for k, v := range t {
				if f, ok := toFloat(v); ok && f >= 0 {
					avg[k] += f
					count[k]++
				}
			}
		}
		parts := []string{}
		for k, total := range avg {
			if count[k] > 0 {
				parts = append(parts, fmt.Sprintf("%s: %.0f", k, total/float64(count[k])))
			}
		}
		if len(parts) > 0 {
			traitsStr = strings.Join(parts, ", ")
		}
	}

	prompt := fmt.Sprintf(`You are the CoffeeDex - a fun Pokemon-themed coffee journal. Write a Pokedex-style entry assigning %s (%s) to this coffee.

COFFEE:
- Name: %s
- Origin: %s
- Roaster: %s
- Variety: %s
- Roast Level: %s
- Processing: %s

TASTING NOTES: %s
FLAVOR SCORES (averaged): %s

Write 2-4 sentences that:
- Reference specific tasting notes from the brews (if any were logged)
- Connect the coffee's character to this specific Pokemon's lore, abilities, appearance, or anime moments
- Are playful and entertaining, not dry or technical
- Make the reader smile and feel like this Pokemon truly fits the coffee

If no tasting notes were logged, base it on the coffee's origin, processing method, and what you'd expect from that profile — but be specific to THIS Pokemon.

Respond with ONLY the description text, no JSON, no quotes, no preamble.`,
		c.PokemonName, c.PokemonType,
		c.CoffeeName, c.Origin, c.Roaster, c.Variety, c.RoastLevel, c.Processing,
		notesStr, traitsStr)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.9,
	})

	req, err := http.NewRequest("POST", openRouterURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/coffee-dex")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return strings.TrimSpace(apiResp.Choices[0].Message.Content), nil
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	}
	return 0, false
}
