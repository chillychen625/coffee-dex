import { Coffee, CoffeePokemon, Pokemon } from "../types/pokemon";

const API_BASE_URL = "http://localhost:8080";

export class CoffeeDexAPI {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  // Coffee endpoints
  async getCoffees(): Promise<Coffee[]> {
    const response = await fetch(`${this.baseUrl}/coffees`);
    if (!response.ok) {
      throw new Error(`Failed to fetch coffees: ${response.statusText}`);
    }
    return response.json();
  }

  async getRecentCoffees(): Promise<Coffee[]> {
    const response = await fetch(`${this.baseUrl}/coffees/recent`);
    if (!response.ok) {
      throw new Error(`Failed to fetch recent coffees: ${response.statusText}`);
    }
    return response.json();
  }

  async getCoffee(id: string): Promise<Coffee> {
    const response = await fetch(`${this.baseUrl}/coffees/${id}`);
    if (!response.ok) {
      throw new Error(`Failed to fetch coffee: ${response.statusText}`);
    }
    return response.json();
  }

  async createCoffee(coffee: Partial<Coffee>): Promise<Coffee> {
    const response = await fetch(`${this.baseUrl}/coffees`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(coffee),
    });
    if (!response.ok) {
      throw new Error(`Failed to create coffee: ${response.statusText}`);
    }
    return response.json();
  }

  async createBrewEntry(coffee: Partial<Coffee>): Promise<Coffee> {
    // Same as createCoffee but used for subsequent brews
    const response = await fetch(`${this.baseUrl}/coffees`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(coffee),
    });
    if (!response.ok) {
      throw new Error(`Failed to create brew entry: ${response.statusText}`);
    }
    return response.json();
  }

  async updateCoffee(id: string, coffee: Partial<Coffee>): Promise<Coffee> {
    const response = await fetch(`${this.baseUrl}/coffees/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(coffee),
    });
    if (!response.ok) {
      throw new Error(`Failed to update coffee: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteCoffee(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/coffees/${id}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to delete coffee: ${response.statusText}`);
    }
  }

  // Pokemon endpoints
  async generatePokemon(coffeeId: string): Promise<CoffeePokemon> {
    const response = await fetch(`${this.baseUrl}/pokemon/${coffeeId}`, {
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(`Failed to generate Pokemon: ${response.statusText}`);
    }
    return response.json();
  }

  async getCoffeePokemon(coffeeId: string): Promise<CoffeePokemon> {
    const response = await fetch(`${this.baseUrl}/pokemon/${coffeeId}`);
    if (!response.ok) {
      throw new Error(
        `Failed to fetch Pokemon for coffee: ${response.statusText}`
      );
    }
    return response.json();
  }

  async updatePokemonNickname(
    coffeeId: string,
    nickname: string
  ): Promise<CoffeePokemon> {
    const response = await fetch(
      `${this.baseUrl}/pokemon/${coffeeId}/nickname`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ nickname }),
      }
    );
    if (!response.ok) {
      throw new Error(`Failed to update nickname: ${response.statusText}`);
    }
    return response.json();
  }

  async getPokedex(): Promise<CoffeePokemon[]> {
    const response = await fetch(`${this.baseUrl}/pokedex`);
    if (!response.ok) {
      throw new Error(`Failed to fetch Pokedex: ${response.statusText}`);
    }
    return response.json();
  }

  async getPokedexStats(): Promise<{
    total_pokemon: number;
    unique_pokemon: number;
    completion_percentage: number;
  }> {
    const response = await fetch(`${this.baseUrl}/pokedex/stats`);
    if (!response.ok) {
      throw new Error(`Failed to fetch Pokedex stats: ${response.statusText}`);
    }
    return response.json();
  }

  async getStatistics(): Promise<{
    total_coffees: number;
    total_pokemon: number;
    completion_percentage: number;
    average_rating: number;
    highest_rated_coffee: { name: string; rating: number } | null;
    lowest_rated_coffee: { name: string; rating: number } | null;
    type_distribution: { [key: string]: number };
    top_origins: Array<{ origin: string; count: number; avg_rating: number }>;
    processing_methods: { [key: string]: number };
    roast_levels: { [key: string]: number };
    trait_averages: { [key: string]: number };
    brewer_stats: Array<{
      brewer: string;
      count: number;
      avg_rating: number;
      avg_brew_time: number;
    }>;
    confidence_metrics: {
      average_confidence: number;
      high_confidence_count: number;
      medium_confidence_count: number;
      low_confidence_count: number;
    };
  }> {
    const response = await fetch(`${this.baseUrl}/statistics`);
    if (!response.ok) {
      throw new Error(`Failed to fetch statistics: ${response.statusText}`);
    }
    return response.json();
  }

  // Brewer endpoints
  async getBrewers(): Promise<
    Array<{
      id: string;
      name: string;
      pokeball_type: string;
      recipes: Array<{
        id: string;
        name: string;
        steps: string[];
      }>;
      created_at: string;
    }>
  > {
    const response = await fetch(`${this.baseUrl}/brewers`);
    if (!response.ok) {
      throw new Error(`Failed to fetch brewers: ${response.statusText}`);
    }
    return response.json();
  }

  async getBrewersWithRecipes(): Promise<
    Array<{
      brewer: {
        id: string;
        name: string;
        pokeball_type: string;
        created_at: string;
      };
      recipes: Coffee[];
    }>
  > {
    const response = await fetch(`${this.baseUrl}/brewers/with-recipes`);
    if (!response.ok) {
      throw new Error(
        `Failed to fetch brewers with recipes: ${response.statusText}`
      );
    }
    return response.json();
  }

  async createBrewer(
    name: string,
    pokeballType: string
  ): Promise<{
    id: string;
    name: string;
    pokeball_type: string;
    created_at: string;
  }> {
    const response = await fetch(`${this.baseUrl}/brewers`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, pokeball_type: pokeballType }),
    });
    if (!response.ok) {
      throw new Error(`Failed to create brewer: ${response.statusText}`);
    }
    return response.json();
  }

  async deleteBrewer(brewerId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/brewers/${brewerId}`, {
      method: "DELETE",
    });
    if (!response.ok) {
      throw new Error(`Failed to delete brewer: ${response.statusText}`);
    }
  }

  async addRecipeToBrewer(brewerId: string, coffeeId: string): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/brewers/${brewerId}/recipes`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ coffee_id: coffeeId }),
      }
    );
    if (!response.ok) {
      throw new Error(`Failed to add recipe to brewer: ${response.statusText}`);
    }
  }

  async removeRecipeFromBrewer(
    brewerId: string,
    coffeeId: string
  ): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/brewers/${brewerId}/recipes/${coffeeId}`,
      {
        method: "DELETE",
      }
    );
    if (!response.ok) {
      throw new Error(
        `Failed to remove recipe from brewer: ${response.statusText}`
      );
    }
  }

  async getPokeballTypes(): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/brewers/pokeball-types`);
    if (!response.ok) {
      throw new Error(`Failed to fetch pokeball types: ${response.statusText}`);
    }
    return response.json();
  }

  async addStandaloneRecipe(
    brewerId: string,
    name: string,
    steps: string[]
  ): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/brewers/${brewerId}/standalone-recipes`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name, steps }),
      }
    );
    if (!response.ok) {
      throw new Error(
        `Failed to add standalone recipe: ${response.statusText}`
      );
    }
  }

  async removeStandaloneRecipe(
    brewerId: string,
    recipeId: string
  ): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/brewers/${brewerId}/standalone-recipes/${recipeId}`,
      {
        method: "DELETE",
      }
    );
    if (!response.ok) {
      throw new Error(
        `Failed to remove standalone recipe: ${response.statusText}`
      );
    }
  }

  // Health check
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      return response.ok;
    } catch (error) {
      return false;
    }
  }
}

// Export a singleton instance
export const api = new CoffeeDexAPI();
