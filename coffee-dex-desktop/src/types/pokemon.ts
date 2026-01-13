export interface Pokemon {
  id: number;
  name: string;
  type: string;
  sprite_path: string;
  base_stats: PokemonStats;
  description: string;
}

export interface PokemonStats {
  hp: number;
  attack: number;
  defense: number;
  speed: number;
  special: number;
}

export interface CoffeePokemon {
  id: string;
  coffee_id: string;
  pokemon_id: number;
  pokemon_name: string;
  nickname?: string;
  level: number;
  mapping_confidence: number;
  llm_description: string;
  trait_mapping: TraitMapping[];
  created_at: string;
}

export interface TraitMapping {
  trait: string;
  pokemon_stat: string;
  reasoning: string;
}

// RoastLevel type
export type RoastLevel =
  | "light"
  | "medium"
  | "dark"
  | "light medium"
  | "medium dark"
  | "unclear";

// ProcessingMethod type
export type ProcessingMethod =
  | "washed"
  | "natural"
  | "honey"
  | "coferment"
  | "experimental";

// Coffee - bean info only
export interface Coffee {
  id: string;
  name: string;
  origin: string;
  roaster: string;
  variety: string;
  roast_level: RoastLevel;
  processing_method: ProcessingMethod;
  created_at: string;
  updated_at: string;
}

// Brew - per-tasting evaluation data
export interface Brew {
  id: string;
  coffee_id: string;
  tasting_notes: [string, string, string, string, string]; // Fixed array of 5 strings
  tasting_traits: TastingTraits;
  rating: number; // 0-10
  recipe: string[];
  dripper: string;
  end_time: DrawDownTime;
  created_at: string;
}

// Coffee with brew stats for list view
export interface CoffeeWithBrewStats extends Coffee {
  brew_count: number;
  average_rating: number;
  can_generate_pokemon: boolean;
  has_pokemon: boolean;
}

// Brew progress for Pokemon generation
export interface BrewProgress {
  count: number;
  required: number;
  can_generate_pokemon: boolean;
  has_pokemon: boolean;
}

export interface DrawDownTime {
  minutes: number;
  seconds: number;
}

export interface TastingTraits {
  berry_intensity: number; // 0-10
  stonefruit_intensity: number; // 0-10
  roast_intensity: number; // 0-10
  citrus_fruits_intensity: number; // 0-10
  bitterness: number; // 0-10
  florality: number; // 0-10
  spice: number; // 0-10
  sweetness: number; // 0-10
  aromatic_intensity: number; // 0-10
  savory: number; // 0-10
  body: number; // 0-10
  cleanliness: number; // 0-10
}
