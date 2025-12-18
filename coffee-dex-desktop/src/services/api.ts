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
