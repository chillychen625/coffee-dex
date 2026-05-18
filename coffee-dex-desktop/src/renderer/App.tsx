import React, { useState, useEffect } from "react";
import { api } from "../services/api";
import {
  Coffee,
  CoffeePokemon,
  TastingTraits,
  Brew,
  BrewWithCoffee,
  BrewProgress,
} from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import CoffeeForm from "./CoffeeForm";
import BrewForm from "./BrewForm";
import Statistics from "./Statistics";
import SpecialItems from "./SpecialItems";
import TitleBar from "../components/TitleBar";
import FlavorWheel from "./FlavorWheel";
import GymBadges from "./GymBadges";
import CoffeeComparison from "./CoffeeComparison";
import LoadingSpinner from "./LoadingSpinner";
import EmptyState from "./EmptyState";
import FormDecoSprites from "./FormDecoSprites";
import BrewChart from "./BrewChart";

// Pokemon type color map for UI accents
const POKEMON_TYPE_COLORS: { [key: string]: string } = {
  Normal: "#a8a878",
  Fire: "#f08030",
  Water: "#6890f0",
  Grass: "#78c850",
  Electric: "#f8d030",
  Ice: "#98d8d8",
  Fighting: "#c03028",
  Poison: "#a040a0",
  Ground: "#e0c068",
  Flying: "#a890f0",
  Psychic: "#f85888",
  Bug: "#a8b820",
  Rock: "#b8a038",
  Ghost: "#705898",
  Dragon: "#7038f8",
  Dark: "#705848",
  Steel: "#b8b8d0",
  Fairy: "#ee99ac",
};

// Get color for a Pokemon type string (handles dual types like "Fire/Water")
const getTypeColor = (pokemonType: string): string => {
  if (!pokemonType) return "#a8a878";
  const primaryType = pokemonType.split("/")[0].trim();
  // Try exact match first, then case-insensitive
  return (
    POKEMON_TYPE_COLORS[primaryType] ||
    POKEMON_TYPE_COLORS[primaryType.charAt(0).toUpperCase() + primaryType.slice(1).toLowerCase()] ||
    "#a8a878"
  );
};

// Get secondary type color for dual types
const getSecondaryTypeColor = (pokemonType: string): string | null => {
  if (!pokemonType || !pokemonType.includes("/")) return null;
  const secondaryType = pokemonType.split("/")[1].trim();
  return (
    POKEMON_TYPE_COLORS[secondaryType] ||
    POKEMON_TYPE_COLORS[secondaryType.charAt(0).toUpperCase() + secondaryType.slice(1).toLowerCase()] ||
    null
  );
};

// Encouraging messages for brew success screen
const BREW_MESSAGES = [
  "Great brew, trainer!",
  "Another fine cup!",
  "Your skills are improving!",
  "Excellent technique!",
  "A brew worthy of a Champion!",
  "Prof. Oak would be proud!",
  "Your Pokemon senses a great brew!",
  "Keep brewing, trainer!",
];

const getRandomBrewMessage = (): string => {
  return BREW_MESSAGES[Math.floor(Math.random() * BREW_MESSAGES.length)];
};

interface AppState {
  view:
    | "start"
    | "home"
    | "coffee-form"
    | "coffee-list"
    | "coffee-detail"
    | "brew-form"
    | "pokedex"
    | "settings"
    | "statistics"
    | "special-items"
    | "brew-success"
    | "recent-brews"
    | "comparison"
    | "pokemon-encounter";
  coffees: Coffee[];
  recentBrewsWithCoffee: BrewWithCoffee[];
  recentCoffees: Coffee[];
  currentCoffee: Coffee | null;
  currentPokemon: CoffeePokemon | null;
  currentBrews: Brew[];
  brewProgress: BrewProgress | null;
  pokedex: CoffeePokemon[];
  pokedexAll: CoffeePokemon[]; // Full unfiltered list
  currentPokedexIndex: number;
  loading: boolean;
  error: string | null;
  backendConnected: boolean;
  formStep: number;
  colorTheme: "red" | "blue" | "yellow";
  pokedexSort: "pokedex" | "date" | "name" | "confidence" | "roast_date";
  pokedexFilters: {
    roaster: string;
    origin: string;
    variety: string;
    roast_level: string;
    processing_method: string;
    pokemon_type: string;
  };
  pokedexFiltersOpen: boolean;
  justCreatedPokemon: boolean;
  promptFirstBrew: boolean; // After creating coffee, prompt to add first brew
  // Brew success screen state
  brewSuccessMessage: string;
  // Pokemon encounter animation state
  encounterPhase: "flash" | "text" | "sprite" | "done";
  encounterPokemon: CoffeePokemon | null;
  // Dashboard stats
  dashboardStats: {
    totalCoffees: number;
    totalBrews: number;
    totalPokemon: number;
    recentPokemon: CoffeePokemon | null;
    featuredCoffee: Coffee | null;
    featuredCoffeeBrews: Brew[];
  } | null;
}

const defaultBrewFormData = (): Partial<Brew> => ({
  coffee_id: "",
  tasting_notes: ["", "", "", "", ""],
  rating: 5,
  recipe: [],
  dripper: "",
  end_time: { minutes: 0, seconds: 0 },
  tasting_traits: {
    berry_intensity: -1,
    stonefruit_intensity: -1,
    roast_intensity: -1,
    citrus_fruits_intensity: -1,
    bitterness: -1,
    florality: -1,
    spice: -1,
    sweetness: -1,
    aromatic_intensity: -1,
    savory: -1,
    body: -1,
    cleanliness: -1,
  } as TastingTraits,
});

// Helper function to calculate days off roast from a roast_date string
const getDaysOffRoast = (roastDate?: string): number => {
  if (!roastDate) return -1;
  const roast = new Date(roastDate);
  const now = new Date();
  const diffTime = now.getTime() - roast.getTime();
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
  return diffDays;
};

// Format days off roast for display
const formatDaysOffRoast = (days: number): string => {
  if (days < 0) return "";
  if (days === 0) return "Roast day";
  if (days === 1) return "1 day off roast";
  return `${days} days off roast`;
};

// Check if coffee is stale (over 6 weeks / 42 days)
const isStale = (days: number): boolean => days >= 42;

// Helper to average a single trait across brews, skipping -1 (not scored)
const averageTraitValue = (brews: Brew[], key: keyof TastingTraits): number => {
  let sum = 0, count = 0;
  brews.forEach((brew) => {
    const v = brew.tasting_traits[key];
    if (v !== undefined && v >= 0) {
      sum += v;
      count++;
    }
  });
  return count === 0 ? -1 : Math.round(sum / count);
};

// Helper to compute average traits from an array of brews
// Traits with value -1 (not scored) are excluded from averages
const computeAverageTraits = (brews: Brew[]): TastingTraits | null => {
  if (brews.length === 0) return null;

  const traitKeys: (keyof TastingTraits)[] = [
    "berry_intensity", "stonefruit_intensity", "roast_intensity",
    "citrus_fruits_intensity", "bitterness", "florality", "spice",
    "sweetness", "aromatic_intensity", "savory", "body", "cleanliness",
  ];

  const result = {} as TastingTraits;
  traitKeys.forEach((key) => {
    result[key] = averageTraitValue(brews, key);
  });
  return result;
};

const App: React.FC = () => {
  const [showBadgeInfo, setShowBadgeInfo] = useState(() => {
    return localStorage.getItem("badgeInfoDismissed") !== "true";
  });

  const dismissBadgeInfo = () => {
    localStorage.setItem("badgeInfoDismissed", "true");
    setShowBadgeInfo(false);
  };

  const [state, setState] = useState<AppState>({
    view: "start",
    coffees: [],
    recentBrewsWithCoffee: [],
    recentCoffees: [],
    currentCoffee: null,
    currentPokemon: null,
    currentBrews: [],
    brewProgress: null,
    pokedex: [],
    pokedexAll: [],
    currentPokedexIndex: 0,
    loading: false,
    error: null,
    backendConnected: false,
    formStep: 1,
    colorTheme: "blue",
    pokedexSort: "pokedex",
    pokedexFilters: {
      roaster: "",
      origin: "",
      variety: "",
      roast_level: "",
      processing_method: "",
      pokemon_type: "",
    },
    pokedexFiltersOpen: false,
    justCreatedPokemon: false,
    promptFirstBrew: false,
    brewSuccessMessage: "",
    encounterPhase: "flash" as const,
    encounterPokemon: null,
    dashboardStats: null,
  });

  const [coffeeFormData, setCoffeeFormData] = useState<Partial<Coffee>>({
    name: "",
    origin: "",
    roaster: "",
    variety: "",
    roast_level: "medium",
    processing_method: "washed",
  });

  const [brewFormData, setBrewFormData] = useState<Partial<Brew>>(
    defaultBrewFormData()
  );

  // Check backend connection on mount
  useEffect(() => {
    checkBackend();
  }, []);

  // Keyboard shortcuts for Pokedex navigation
  useEffect(() => {
    if (state.view !== "pokedex") return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
        e.preventDefault();
        if (e.key === "ArrowLeft") {
          navigatePokedex("prev");
        } else if (e.key === "ArrowRight") {
          navigatePokedex("next");
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [state.view, state.currentPokedexIndex, state.pokedex.length]);

  // Auto-advance encounter animation phases
  useEffect(() => {
    if (state.view !== "pokemon-encounter") return;

    const timers: NodeJS.Timeout[] = [];

    if (state.encounterPhase === "flash") {
      timers.push(setTimeout(() => {
        setState((prev) => ({ ...prev, encounterPhase: "text" }));
      }, 800));
    } else if (state.encounterPhase === "text") {
      timers.push(setTimeout(() => {
        setState((prev) => ({ ...prev, encounterPhase: "sprite" }));
      }, 1200));
    } else if (state.encounterPhase === "sprite") {
      timers.push(setTimeout(() => {
        setState((prev) => ({ ...prev, encounterPhase: "done" }));
      }, 1000));
    }

    return () => timers.forEach(t => clearTimeout(t));
  }, [state.view, state.encounterPhase]);

  const checkBackend = async () => {
    const maxRetries = 30;
    const retryDelay = 500; // ms

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      const connected = await api.healthCheck();
      if (connected) {
        setState((prev) => ({ ...prev, backendConnected: true, error: null }));
        return;
      }
      // Wait before next retry
      await new Promise((resolve) => setTimeout(resolve, retryDelay));
    }

    // All retries exhausted
    setState((prev) => ({
      ...prev,
      backendConnected: false,
      error: "Backend not connected. Please start the server.",
    }));
  };

  const loadCoffees = async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const coffees = await api.getCoffees();
      setState((prev) => ({
        ...prev,
        coffees,
        loading: false,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to load coffees: ${error}`,
        loading: false,
      }));
    }
  };

  const loadCoffeeDetail = async (coffeeId: string) => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const [coffee, brews, progress] = await Promise.all([
        api.getCoffee(coffeeId),
        api.getBrewsForCoffee(coffeeId),
        api.getBrewProgress(coffeeId),
      ]);

      // Try to load Pokemon if exists
      let pokemon: CoffeePokemon | null = null;
      if (progress.has_pokemon) {
        try {
          pokemon = await api.getCoffeePokemon(coffeeId);
        } catch {
          // No Pokemon yet
        }
      }

      setState((prev) => ({
        ...prev,
        currentCoffee: coffee,
        currentBrews: brews,
        brewProgress: progress,
        currentPokemon: pokemon,
        loading: false,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to load coffee details: ${error}`,
        loading: false,
      }));
    }
  };

  const loadPokedex = async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const [allPokedex, coffees] = await Promise.all([
        api.getPokedex(),
        api.getCoffees(),
      ]);

      const filtered = filterPokedex(allPokedex, state.pokedexFilters, coffees);
      const sorted = sortPokedex(filtered, state.pokedexSort, coffees);

      if (sorted.length > 0) {
        const firstPokemon = sorted[0];
        const [coffee, brews] = await Promise.all([
          api.getCoffee(firstPokemon.coffee_id),
          api.getBrewsForCoffee(firstPokemon.coffee_id),
        ]);
        setState((prev) => ({
          ...prev,
          pokedexAll: allPokedex,
          pokedex: sorted,
          coffees,
          currentPokemon: firstPokemon,
          currentCoffee: coffee,
          currentBrews: brews || [],
          currentPokedexIndex: 0,
          loading: false,
        }));
      } else {
        setState((prev) => ({
          ...prev,
          pokedexAll: allPokedex,
          pokedex: sorted,
          coffees,
          loading: false,
        }));
      }
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to load Pokedex: ${error}`,
        loading: false,
      }));
    }
  };

  type SortOption = "pokedex" | "date" | "name" | "confidence" | "roast_date";

  const filterPokedex = (
    allPokedex: CoffeePokemon[],
    filters: AppState["pokedexFilters"],
    coffees: Coffee[]
  ): CoffeePokemon[] => {
    const hasActiveFilter = Object.values(filters).some((v) => v !== "");
    if (!hasActiveFilter) return allPokedex;

    const coffeeMap = new Map<string, Coffee>();
    coffees.forEach((c) => coffeeMap.set(c.id, c));

    return allPokedex.filter((entry) => {
      const coffee = coffeeMap.get(entry.coffee_id);
      if (!coffee) return false;

      if (filters.roaster && coffee.roaster !== filters.roaster) return false;
      if (filters.origin && coffee.origin !== filters.origin) return false;
      if (filters.variety && coffee.variety !== filters.variety) return false;
      if (filters.roast_level && coffee.roast_level !== filters.roast_level) return false;
      if (filters.processing_method && coffee.processing_method !== filters.processing_method) return false;
      if (filters.pokemon_type && !entry.pokemon_type.includes(filters.pokemon_type)) return false;

      return true;
    });
  };

  const sortPokedex = (
    pokedex: CoffeePokemon[],
    sortBy: SortOption,
    coffees?: Coffee[]
  ) => {
    const sorted = [...pokedex];
    switch (sortBy) {
      case "pokedex":
        return sorted.sort((a, b) => a.pokemon_id - b.pokemon_id);
      case "date":
        return sorted.reverse();
      case "name":
        return sorted.sort((a, b) =>
          a.pokemon_name.localeCompare(b.pokemon_name)
        );
      case "confidence":
        return sorted.sort(
          (a, b) => b.mapping_confidence - a.mapping_confidence
        );
      case "roast_date": {
        const roastDateMap = new Map<string, string>();
        if (coffees) {
          coffees.forEach((c) => {
            if (c.roast_date) {
              roastDateMap.set(c.id, c.roast_date);
            }
          });
        }
        return sorted.sort((a, b) => {
          const dateA = roastDateMap.get(a.coffee_id) || "";
          const dateB = roastDateMap.get(b.coffee_id) || "";
          if (!dateA && !dateB) return 0;
          if (!dateA) return 1;
          if (!dateB) return -1;
          return dateB.localeCompare(dateA);
        });
      }
      default:
        return sorted;
    }
  };

  const applyFiltersAndSort = async (
    allPokedex: CoffeePokemon[],
    filters: AppState["pokedexFilters"],
    sortBy: SortOption,
    coffees: Coffee[]
  ) => {
    const filtered = filterPokedex(allPokedex, filters, coffees);
    const sorted = sortPokedex(filtered, sortBy, coffees);

    if (sorted.length > 0) {
      try {
        const firstPokemon = sorted[0];
        const [coffee, brews] = await Promise.all([
          api.getCoffee(firstPokemon.coffee_id),
          api.getBrewsForCoffee(firstPokemon.coffee_id),
        ]);
        setState((prev) => ({
          ...prev,
          pokedex: sorted,
          currentPokemon: firstPokemon,
          currentCoffee: coffee,
          currentBrews: brews || [],
          currentPokedexIndex: 0,
          loading: false,
        }));
      } catch (error) {
        setState((prev) => ({
          ...prev,
          error: `Failed to load coffee: ${error}`,
          loading: false,
        }));
      }
    } else {
      setState((prev) => ({
        ...prev,
        pokedex: sorted,
        currentPokemon: null,
        currentCoffee: null,
        currentBrews: [],
        currentPokedexIndex: 0,
        loading: false,
      }));
    }
  };

  const handleSortChange = async (sortBy: SortOption) => {
    setState((prev) => ({ ...prev, pokedexSort: sortBy, loading: true }));
    await applyFiltersAndSort(state.pokedexAll, state.pokedexFilters, sortBy, state.coffees);
  };

  const handleFilterChange = async (filterKey: string, value: string) => {
    const newFilters = { ...state.pokedexFilters, [filterKey]: value };
    setState((prev) => ({ ...prev, pokedexFilters: newFilters, loading: true }));
    await applyFiltersAndSort(state.pokedexAll, newFilters, state.pokedexSort, state.coffees);
  };

  const navigatePokedex = async (direction: "prev" | "next") => {
    const newIndex =
      direction === "next"
        ? Math.min(state.currentPokedexIndex + 1, state.pokedex.length - 1)
        : Math.max(state.currentPokedexIndex - 1, 0);

    if (newIndex !== state.currentPokedexIndex) {
      setState((prev) => ({ ...prev, loading: true }));
      try {
        const pokemon = state.pokedex[newIndex];
        const [coffee, brews] = await Promise.all([
          api.getCoffee(pokemon.coffee_id),
          api.getBrewsForCoffee(pokemon.coffee_id),
        ]);
        setState((prev) => ({
          ...prev,
          currentPokemon: pokemon,
          currentCoffee: coffee,
          currentBrews: brews || [],
          currentPokedexIndex: newIndex,
          loading: false,
        }));
      } catch (error) {
        setState((prev) => ({
          ...prev,
          error: `Failed to load coffee: ${error}`,
          loading: false,
        }));
      }
    }
  };

  const loadRecentCoffees = async () => {
    try {
      const recent = await api.getRecentCoffees();
      setState((prev) => ({ ...prev, recentCoffees: recent }));
    } catch (error) {
      console.error("Failed to load recent coffees:", error);
    }
  };

  const loadDashboardStats = async () => {
    try {
      const [coffees, pokedex, recentBrews] = await Promise.all([
        api.getCoffees(),
        api.getPokedex(),
        api.getRecentBrewsWithCoffee(),
      ]);

      // Count total brews from all recent + any other source
      // Use the statistics endpoint for accurate counts
      let totalBrews = 0;
      let featuredCoffee: Coffee | null = null;
      let featuredCoffeeBrews: Brew[] = [];
      try {
        const stats = await api.getStatistics();
        totalBrews = recentBrews.length; // Use recent brews count as fallback
        // Try to get total from stats if available
        if (stats && typeof stats.total_coffees === "number") {
          // stats.total_coffees is available; we'll use coffees.length for actual count
        }
      } catch {
        totalBrews = recentBrews.length;
      }

      // Get recent Pokemon (most recently created)
      const recentPokemon = pokedex.length > 0 ? pokedex[pokedex.length - 1] : null;

      // Featured coffee: most recently brewed, or highest rated
      if (recentBrews.length > 0) {
        const latestBrewCoffeeId = recentBrews[0].coffee_id;
        try {
          featuredCoffee = await api.getCoffee(latestBrewCoffeeId);
          featuredCoffeeBrews = await api.getBrewsForCoffee(latestBrewCoffeeId);
        } catch {
          // fallback
        }
      } else if (coffees.length > 0) {
        featuredCoffee = coffees[coffees.length - 1];
        try {
          featuredCoffeeBrews = await api.getBrewsForCoffee(featuredCoffee.id);
        } catch {
          // fallback
        }
      }

      // Count all brews across all coffees
      let brewCount = 0;
      for (const coffee of coffees) {
        try {
          const progress = await api.getBrewProgress(coffee.id);
          brewCount += progress.count;
        } catch {
          // skip
        }
      }

      setState((prev) => ({
        ...prev,
        coffees,
        pokedex,
        recentBrewsWithCoffee: recentBrews,
        dashboardStats: {
          totalCoffees: coffees.length,
          totalBrews: brewCount || recentBrews.length,
          totalPokemon: pokedex.length,
          recentPokemon,
          featuredCoffee,
          featuredCoffeeBrews,
        },
      }));
    } catch (error) {
      console.error("Failed to load dashboard stats:", error);
    }
  };

  const resetCoffeeForm = () => {
    setCoffeeFormData({
      name: "",
      origin: "",
      roaster: "",
      variety: "",
      roast_level: "medium",
      processing_method: "washed",
    });
    setState((prev) => ({
      ...prev,
      formStep: 1,
      error: null,
    }));
  };

  const resetBrewForm = () => {
    setBrewFormData(defaultBrewFormData());
    setState((prev) => ({
      ...prev,
      formStep: 1,
      error: null,
    }));
  };

  const handleCoffeeSubmit = async () => {
    if (!coffeeFormData.name || !coffeeFormData.origin) {
      setState((prev) => ({
        ...prev,
        error: "Please fill in required fields (name, origin)",
      }));
      return;
    }

    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const newCoffee = await api.createCoffee(coffeeFormData);
      setState((prev) => ({
        ...prev,
        currentCoffee: newCoffee,
        loading: false,
        promptFirstBrew: true,
        view: "coffee-detail",
      }));
      resetCoffeeForm();
      await loadCoffeeDetail(newCoffee.id);
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to create coffee: ${error}`,
        loading: false,
      }));
    }
  };

  const handleBrewSubmit = async () => {
    if (!state.currentCoffee) return;

    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const brewData = {
        ...brewFormData,
        coffee_id: state.currentCoffee.id,
      };
      await api.createBrew(brewData);

      // Reload coffee detail to update brew count and progress
      await loadCoffeeDetail(state.currentCoffee.id);
      resetBrewForm();
      setState((prev) => ({
        ...prev,
        loading: false,
        view: "brew-success",
        promptFirstBrew: false,
        brewSuccessMessage: getRandomBrewMessage(),
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to create brew: ${error}`,
        loading: false,
      }));
    }
  };

  const handleGeneratePokemon = async () => {
    if (!state.currentCoffee) return;

    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const pokemon = await api.generatePokemon(state.currentCoffee.id);
      // Show the encounter animation instead of going straight to pokedex
      setState((prev) => ({
        ...prev,
        currentPokemon: pokemon,
        encounterPokemon: pokemon,
        encounterPhase: "flash",
        view: "pokemon-encounter",
        loading: false,
        justCreatedPokemon: true,
      }));
      // Reload pokedex in background
      loadPokedex();
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to generate Pokemon: ${error}`,
        loading: false,
      }));
    }
  };

  const handleMarkAsFinished = async () => {
    if (!state.currentCoffee) return;

    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const updatedCoffee = await api.markCoffeeAsFinished(state.currentCoffee.id);
      // Reload brew progress to get updated can_generate_pokemon status
      const progress = await api.getBrewProgress(state.currentCoffee.id);
      setState((prev) => ({
        ...prev,
        currentCoffee: updatedCoffee,
        brewProgress: progress,
        loading: false,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to mark coffee as finished: ${error}`,
        loading: false,
      }));
    }
  };

  const renderStart = () => (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ position: "relative", justifyContent: "center" }}
      >
        <FormDecoSprites seed="start-screen" spin={true} />
        <div style={{ textAlign: "center", position: "relative", zIndex: 1 }}>
          <h1
            className="pokemon-title"
            style={{ fontSize: "28px", marginBottom: "24px" }}
          >
            COFFEEDEX
          </h1>
          <div style={{ fontSize: "11px", marginBottom: "8px", fontStyle: "italic" }}>
            Gotta Brew 'Em All!
          </div>
          <div
            className="pokemon-textbox"
            style={{
              fontSize: "10px",
              marginBottom: "32px",
              padding: "12px",
              lineHeight: "1.5",
            }}
          >
            Your coffee journey awaits! Log your brews, discover Pokemon partners, and become the ultimate Coffee Master!
          </div>
          <button
            className="pokemon-button"
            onClick={() => {
              checkBackend();
              loadDashboardStats();
              setState((prev) => ({ ...prev, view: "home" }));
            }}
            style={{ fontSize: "16px", padding: "16px 48px" }}
          >
            Press Start
          </button>
          <div style={{ marginTop: "16px", fontSize: "9px", opacity: 0.7 }}>
            151 Pokemon to collect!
          </div>
        </div>
      </div>
    </div>
  );

  const renderHome = () => {
    const stats = state.dashboardStats;
    const completedIds = new Set(state.pokedex.map(p => p.coffee_id));
    const openBagCount = state.coffees.filter(c => !completedIds.has(c.id) && !c.is_finished).length;

    // Reusable menu item component
    const MenuItem: React.FC<{
      label: string;
      subtitle: string;
      badge?: string;
      onClick: () => void;
      disabled?: boolean;
      accent?: string;
    }> = ({ label, subtitle, badge, onClick, disabled, accent }) => (
      <button
        className="pokemon-textbox"
        onClick={onClick}
        disabled={disabled}
        style={{
          width: "100%",
          textAlign: "left",
          cursor: disabled ? "not-allowed" : "pointer",
          border: `3px solid ${accent || "var(--border-color)"}`,
          padding: "10px 12px",
          display: "flex",
          alignItems: "center",
          gap: "8px",
          opacity: disabled ? 0.5 : 1,
          transition: "background 0.1s",
          marginBottom: "0",
          marginTop: "0",
        }}
      >
        <div style={{ flex: 1 }}>
          <div style={{ fontSize: "13px", fontWeight: "bold", letterSpacing: "1px" }}>
            {label}
          </div>
          <div style={{ fontSize: "10px", opacity: 0.7, marginTop: "2px" }}>
            {subtitle}
          </div>
        </div>
        {badge && (
          <div style={{
            fontSize: "11px",
            fontWeight: "bold",
            padding: "2px 8px",
            background: accent ? `${accent}20` : "rgba(0,0,0,0.08)",
            border: `2px solid ${accent || "var(--border-color)"}`,
            color: accent || "inherit",
            whiteSpace: "nowrap",
          }}>
            {badge}
          </div>
        )}
      </button>
    );

    // Section divider
    const SectionDivider: React.FC<{ label: string }> = ({ label }) => (
      <div style={{
        display: "flex",
        alignItems: "center",
        gap: "8px",
        margin: "12px 0 8px",
        fontSize: "10px",
        fontWeight: "bold",
        letterSpacing: "2px",
        color: "var(--gb-dark)",
        opacity: 0.8,
      }}>
        <div style={{ flex: 1, height: "2px", background: "var(--gb-dark)", opacity: 0.3 }} />
        <span>{label}</span>
        <div style={{ flex: 1, height: "2px", background: "var(--gb-dark)", opacity: 0.3 }} />
      </div>
    );

    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ position: "relative", gap: "4px" }}
        >
          <FormDecoSprites seed="home" spin={true} />

          {/* Compact Header with Stats */}
          <div style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            position: "relative",
            zIndex: 1,
            marginBottom: "4px",
          }}>
            <div>
              <h1 style={{
                fontSize: "18px",
                fontWeight: "bold",
                letterSpacing: "2px",
                textShadow: "2px 2px 0px rgba(0,0,0,0.2)",
                fontFamily: "'Press Start 2P', monospace",
                margin: 0,
              }}>COFFEEDEX</h1>
              <div style={{ fontSize: "10px", opacity: 0.6, marginTop: "2px" }}>
                Gotta Brew 'Em All!
              </div>
            </div>
            {stats && (
              <div style={{
                display: "flex",
                gap: "12px",
                fontSize: "10px",
                textAlign: "center",
              }}>
                <div>
                  <div style={{ fontSize: "16px", fontWeight: "bold" }}>{stats.totalCoffees}</div>
                  <div style={{ opacity: 0.7 }}>Bags</div>
                </div>
                <div>
                  <div style={{ fontSize: "16px", fontWeight: "bold" }}>{stats.totalBrews}</div>
                  <div style={{ opacity: 0.7 }}>Brews</div>
                </div>
                <div>
                  <div style={{ fontSize: "16px", fontWeight: "bold" }}>{stats.totalPokemon}</div>
                  <div style={{ opacity: 0.7 }}>Pkmn</div>
                </div>
              </div>
            )}
          </div>

          {/* Recent Pokemon Catch - compact banner */}
          {stats?.recentPokemon && (
            <div style={{
              display: "flex",
              alignItems: "center",
              gap: "8px",
              padding: "6px 10px",
              background: "rgba(0,0,0,0.05)",
              border: "2px solid var(--gb-dark)",
              fontSize: "11px",
            }}>
              <img
                src={`./pokemon-sprites/animated/${stats.recentPokemon.pokemon_id}.gif`}
                alt={stats.recentPokemon.pokemon_name}
                style={{ width: "32px", height: "32px", imageRendering: "pixelated" }}
                onError={(e) => {
                  e.currentTarget.src = `./pokemon-sprites/${String(stats!.recentPokemon!.pokemon_id).padStart(3, "0")}.png`;
                }}
              />
              <div style={{ flex: 1 }}>
                <span style={{ fontWeight: "bold" }}>Latest: </span>
                {stats.recentPokemon.pokemon_name}
              </div>
              <span style={{ opacity: 0.6 }}>Lv.{(stats.recentPokemon as any).level || "?"}</span>
            </div>
          )}

          {/* Featured Coffee with FlavorWheel */}
          {stats?.featuredCoffee && stats.featuredCoffeeBrews.length > 0 && (() => {
            const avgTraits = computeAverageTraits(stats.featuredCoffeeBrews);
            if (!avgTraits) return null;
            return (
              <div style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                padding: "6px 10px",
                background: "rgba(0,0,0,0.03)",
                border: "2px solid var(--gb-dark)",
                fontSize: "11px",
              }}>
                <FlavorWheel traits={avgTraits} size={64} showLabels={false} />
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: "bold" }}>{stats.featuredCoffee!.name}</div>
                  <div style={{ opacity: 0.7 }}>{stats.featuredCoffee!.origin}</div>
                </div>
              </div>
            );
          })()}

          {!state.backendConnected && (
            <div
              className="pokemon-textbox"
              style={{ background: "#ffcccc", borderColor: "#cc0000", margin: "4px 0" }}
            >
              <div style={{ fontSize: "11px" }}>
                Backend not connected!
                <br />
                Start server: go run main.go -storage=mysql
              </div>
            </div>
          )}

          {/* ── BREWING ── */}
          <SectionDivider label="BREWING" />

          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <MenuItem
              label="NEW COFFEE"
              subtitle="Log a new bag to your collection"
              badge="+"
              accent="#78c850"
              disabled={!state.backendConnected}
              onClick={() => {
                resetCoffeeForm();
                setState((prev) => ({ ...prev, view: "coffee-form", formStep: 1 }));
              }}
            />
            <MenuItem
              label="OPEN BAGS"
              subtitle="Active coffees ready to brew"
              badge={openBagCount > 0 ? String(openBagCount) : undefined}
              disabled={!state.backendConnected}
              onClick={async () => {
                await loadCoffees();
                setState((prev) => ({ ...prev, view: "coffee-list" }));
              }}
            />
            <MenuItem
              label="RECENT BREWS"
              subtitle="Your latest tastings"
              badge={stats ? String(stats.totalBrews) : undefined}
              disabled={!state.backendConnected}
              onClick={async () => {
                try {
                  const recentBrews = await api.getRecentBrewsWithCoffee();
                  setState((prev) => ({ ...prev, recentBrewsWithCoffee: recentBrews, view: "recent-brews" }));
                } catch (err) {
                  console.error("Failed to load recent brews:", err);
                  setState((prev) => ({ ...prev, recentBrewsWithCoffee: [], view: "recent-brews" }));
                }
              }}
            />
          </div>

          {/* ── COLLECTION ── */}
          <SectionDivider label="COLLECTION" />

          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <MenuItem
              label="POKEDEX"
              subtitle="Your Pokemon partners"
              badge={stats ? `${stats.totalPokemon}/151` : undefined}
              accent="#6890f0"
              disabled={!state.backendConnected}
              onClick={() => {
                loadPokedex();
                setState((prev) => ({ ...prev, view: "pokedex" }));
              }}
            />
            <MenuItem
              label="COMPARE"
              subtitle="Side-by-side coffee analysis"
              disabled={!state.backendConnected}
              onClick={() => setState((prev) => ({ ...prev, view: "comparison" }))}
            />
          </div>

          {/* ── TRAINER ── */}
          <SectionDivider label="TRAINER" />

          <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
            <MenuItem
              label="STATISTICS"
              subtitle="Collection analytics & trends"
              disabled={!state.backendConnected}
              onClick={() => setState((prev) => ({ ...prev, view: "statistics" }))}
            />
            <MenuItem
              label="BREWERS"
              subtitle="Your brewing equipment & recipes"
              disabled={!state.backendConnected}
              onClick={() => setState((prev) => ({ ...prev, view: "special-items" }))}
            />
            <MenuItem
              label="SETTINGS"
              subtitle="Theme & preferences"
              disabled={!state.backendConnected}
              onClick={() => setState((prev) => ({ ...prev, view: "settings" }))}
            />
          </div>

          {/* Gym Badges */}
          {state.backendConnected && (
            <div style={{ marginTop: "auto", paddingTop: "8px" }}>
              {showBadgeInfo && (
                <div
                  className="pokemon-textbox"
                  style={{
                    marginBottom: "8px",
                    fontSize: "9px",
                    position: "relative",
                    background: "linear-gradient(135deg, #fffde7 0%, #fff9c4 100%)",
                  }}
                >
                  <button
                    onClick={dismissBadgeInfo}
                    style={{
                      position: "absolute",
                      top: "4px",
                      right: "4px",
                      background: "none",
                      border: "none",
                      cursor: "pointer",
                      fontSize: "12px",
                      padding: "2px 6px",
                    }}
                  >
                    x
                  </button>
                  <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
                    How to Earn Gym Badges:
                  </div>
                  <div style={{ lineHeight: "1.4" }}>
                    Log coffees and brews to unlock badges! Try different origins,
                    processing methods, and collect various Pokemon types.
                    Click any badge to see your progress!
                  </div>
                </div>
              )}
              <GymBadges compact={true} />
            </div>
          )}
        </div>
      </div>
    );
  };

  const renderCoffeeList = () => {
    const allCoffees =
      state.coffees.length > 0 ? state.coffees : state.recentCoffees;

    // Create a set of coffee IDs that have Pokemon (completed bags)
    const completedCoffeeIds = new Set<string>();
    state.pokedex.forEach((pokemon) => {
      completedCoffeeIds.add(pokemon.coffee_id);
    });

    // Filter to only show "open bags" - coffees without Pokemon AND not marked as finished
    const openBags = allCoffees
      .filter(coffee => !completedCoffeeIds.has(coffee.id) && !coffee.is_finished)
      // Sort by roast date (newest first), coffees without roast date go last
      .sort((a, b) => {
        if (!a.roast_date && !b.roast_date) return 0;
        if (!a.roast_date) return 1;
        if (!b.roast_date) return -1;
        return b.roast_date.localeCompare(a.roast_date);
      });

    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ position: "relative" }}
        >
          <FormDecoSprites seed="coffee-list" spin={true} />
          <button
            className="pokemon-button mb-md"
            onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
            style={{ position: "relative", zIndex: 1 }}
          >
            Back
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "14px", position: "relative", zIndex: 1 }}>
            OPEN BAGS
          </h2>

          <div className="pokemon-textbox mb-sm" style={{ fontSize: "8px", textAlign: "center", padding: "4px" }}>
            {openBags.length} open bag{openBags.length !== 1 ? 's' : ''} • {completedCoffeeIds.size} completed (in Pokedex)
          </div>

          {openBags.length === 0 ? (
            <EmptyState
              variant="no-coffees"
              title="No Open Bags"
              message="All your coffees have been completed! Add a new coffee or check out your Pokedex."
            />
          ) : (
            <div>
              {openBags.map((coffee) => {
                const daysOff = getDaysOffRoast(coffee.roast_date);
                const stale = isStale(daysOff);
                return (
                  <button
                    key={coffee.id}
                    className="pokemon-textbox mb-sm"
                    style={{
                      width: "100%",
                      textAlign: "left",
                      cursor: "pointer",
                      border: "2px solid #000",
                    }}
                    onClick={async () => {
                      await loadCoffeeDetail(coffee.id);
                      setState((prev) => ({ ...prev, view: "coffee-detail" }));
                    }}
                  >
                    <div style={{ fontWeight: "bold", fontSize: "11px", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                      <span>{coffee.name}</span>
                      {daysOff >= 0 && (
                        <span style={{
                          fontSize: "8px",
                          fontWeight: "normal",
                          color: stale ? "#cc0000" : "inherit"
                        }}>
                          {daysOff}d{stale && " ⚠"}
                        </span>
                      )}
                    </div>
                    <div style={{ fontSize: "9px" }}>
                      {coffee.origin} | {coffee.roaster}
                    </div>
                    <div style={{ fontSize: "8px", marginTop: "4px" }}>
                      {coffee.roast_level} | {coffee.processing_method}
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </div>
    );
  };

  const renderRecentBrews = () => {
    const recentBrews = state.recentBrewsWithCoffee || [];

    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ position: "relative" }}
        >
          <FormDecoSprites seed="recent-brews" spin={true} />
          <button
            className="pokemon-button mb-md"
            onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
            style={{ position: "relative", zIndex: 1 }}
          >
            Back
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "14px", position: "relative", zIndex: 1 }}>
            RECENT BREWS
          </h2>

          {recentBrews.length === 0 ? (
            <EmptyState
              variant="no-coffees"
              title="No Brews Yet"
              message="Log your first brew from a coffee's detail page!"
            />
          ) : (
            <div>
              {recentBrews.map((brew) => (
                <button
                  key={brew.id}
                  className="pokemon-textbox mb-sm"
                  style={{
                    width: "100%",
                    textAlign: "left",
                    cursor: "pointer",
                    border: "2px solid #000",
                  }}
                  onClick={async () => {
                    await loadCoffeeDetail(brew.coffee_id);
                    setState((prev) => ({ ...prev, view: "coffee-detail" }));
                  }}
                >
                  <div style={{ fontWeight: "bold", fontSize: "11px", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <span>{brew.coffee_name}</span>
                    <span style={{ fontSize: "10px", fontWeight: "normal" }}>
                      {brew.rating}/10
                    </span>
                  </div>
                  <div style={{ fontSize: "9px", marginTop: "2px" }}>
                    {brew.coffee_origin} | {brew.dripper || "No dripper"}
                    {brew.days_off_roast >= 0 && (
                      <span> | {brew.days_off_roast}d off roast</span>
                    )}
                  </div>
                  {brew.tasting_notes.filter((n) => n).length > 0 && (
                    <div style={{ fontSize: "8px", marginTop: "4px", opacity: 0.8 }}>
                      {brew.tasting_notes.filter((n) => n).join(", ")}
                    </div>
                  )}
                  <div style={{ fontSize: "8px", marginTop: "2px", opacity: 0.6, display: "flex", justifyContent: "space-between" }}>
                    <span>
                      {brew.end_time.minutes}:{String(brew.end_time.seconds).padStart(2, "0")}
                    </span>
                    <span>
                      {new Date(brew.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    );
  };

  const renderCoffeeDetail = () => {
    if (!state.currentCoffee || !state.brewProgress) {
      return (
        <div className="pokemon-screen">
          <div
            className="pokemon-frame"
          >
            <LoadingSpinner variant="default" message="Loading coffee..." />
          </div>
        </div>
      );
    }

    const coffee = state.currentCoffee;
    const progress = state.brewProgress;
    const brews = state.currentBrews || [];

    const progressPercent = Math.min(
      (progress.count / progress.required) * 100,
      100
    );
    const canGenerate = progress.can_generate_pokemon && !progress.has_pokemon;

    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{
            position: "relative",
            ...(state.currentPokemon ? {
              borderColor: getTypeColor(state.currentPokemon.pokemon_type),
              boxShadow: `inset -2px -2px 0px rgba(0,0,0,0.2), inset 2px 2px 0px rgba(255,255,255,0.5), 0 0 8px ${getTypeColor(state.currentPokemon.pokemon_type)}30`
            } : {})
          }}
        >
          <FormDecoSprites seed={`detail-${coffee.id}`} spin={true} />
          <button
            className="pokemon-button mb-md"
            onClick={() =>
              setState((prev) => ({
                ...prev,
                view: "coffee-list",
                promptFirstBrew: false,
              }))
            }
            style={{ position: "relative", zIndex: 1 }}
          >
            Back
          </button>

          <h2
            className="pokemon-title"
            style={{ fontSize: "14px", marginBottom: "8px", position: "relative", zIndex: 1 }}
          >
            {coffee.name.toUpperCase()}
          </h2>

          <div className="pokemon-textbox mb-sm" style={{ fontSize: "10px" }}>
            <div>
              <strong>Origin:</strong> {coffee.origin}
            </div>
            <div>
              <strong>Roaster:</strong> {coffee.roaster}
            </div>
            {coffee.variety && (
              <div>
                <strong>Variety:</strong> {coffee.variety}
              </div>
            )}
            <div>
              <strong>Roast:</strong> {coffee.roast_level}
            </div>
            <div>
              <strong>Process:</strong> {coffee.processing_method}
            </div>
            {(() => {
              const daysOff = getDaysOffRoast(coffee.roast_date);
              if (daysOff < 0) return null;
              const stale = isStale(daysOff);
              return (
                <div style={{ marginTop: "4px", color: stale ? "#cc0000" : "inherit" }}>
                  <strong>Age:</strong> {formatDaysOffRoast(daysOff)}{stale && " ⚠"}
                </div>
              );
            })()}
          </div>

          {/* Brew Progress Bar */}
          <div className="pokemon-textbox mb-sm">
            <div
              style={{
                fontWeight: "bold",
                marginBottom: "8px",
                fontSize: "10px",
              }}
            >
              BREW PROGRESS
            </div>
            <div className="pokemon-stat-row">
              <div className="pokemon-stat-bar" style={{ flex: 1 }}>
                <div
                  className={`pokemon-stat-fill ${
                    progress.has_pokemon
                      ? "high"
                      : canGenerate
                      ? "medium"
                      : "low"
                  }`}
                  style={{ width: `${progressPercent}%` }}
                ></div>
              </div>
              <div className="pokemon-stat-value" style={{ fontSize: "10px" }}>
                {progress.count}/{progress.required}
              </div>
            </div>
            {progress.has_pokemon && (
              <div
                style={{
                  fontSize: "9px",
                  marginTop: "4px",
                  color: "#00aa00",
                  fontWeight: "bold",
                }}
              >
                Pokemon Captured!
              </div>
            )}
            {canGenerate && (
              <div
                style={{
                  fontSize: "9px",
                  marginTop: "4px",
                  color: "#aa6600",
                  fontWeight: "bold",
                }}
              >
                Ready to Generate!
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "8px",
              marginBottom: "12px",
            }}
          >
            {/* Primary actions row */}
            <div style={{ display: "flex", gap: "8px" }}>
              <button
                className="pokemon-button"
                style={{ flex: 1, fontSize: "10px", padding: "8px" }}
                onClick={() => {
                  resetBrewForm();
                  setBrewFormData((prev) => ({
                    ...prev,
                    coffee_id: coffee.id,
                  }));
                  setState((prev) => ({
                    ...prev,
                    view: "brew-form",
                    formStep: 1,
                  }));
                }}
              >
                Log Brew
              </button>
              {canGenerate && (
                <button
                  className="pokemon-button"
                  style={{
                    flex: 1,
                    fontSize: "10px",
                    padding: "8px",
                    background: "#ffcc00",
                  }}
                  onClick={handleGeneratePokemon}
                >
                  Generate Pokemon!
                </button>
              )}
              {progress.has_pokemon && state.currentPokemon && (
                <button
                  className="pokemon-button"
                  style={{ flex: 1, fontSize: "10px", padding: "8px" }}
                  onClick={() => {
                    setState((prev) => ({ ...prev, view: "pokedex" }));
                  }}
                >
                  View Pokemon
                </button>
              )}
            </div>

            {/* Mark as Finished button - show for any unfinished coffee without Pokemon */}
            {!coffee.is_finished && !progress.has_pokemon && (
              <button
                className="pokemon-button"
                style={{
                  fontSize: "9px",
                  padding: "6px 8px",
                  background: "#cc9966",
                  opacity: 0.9,
                }}
                onClick={handleMarkAsFinished}
                title="Mark this bag as finished and generate a Pokemon"
              >
                {brews.length > 0
                  ? `Bag Finished? Generate Pokemon from ${brews.length} brew${brews.length !== 1 ? "s" : ""}`
                  : "Close Bag & Generate Pokemon"}
              </button>
            )}

            {/* Show finished badge */}
            {coffee.is_finished && !progress.has_pokemon && (
              <div
                className="pokemon-textbox"
                style={{
                  fontSize: "9px",
                  padding: "4px 8px",
                  background: "#e8e8e8",
                  textAlign: "center",
                }}
              >
                Bag marked as finished - ready to generate Pokemon!
              </div>
            )}
          </div>

          {/* Prompt first brew message */}
          {state.promptFirstBrew && brews.length === 0 && (
            <div
              className="pokemon-textbox mb-sm"
              style={{
                background: "#ffffcc",
                borderColor: "#aaaa00",
                fontSize: "10px",
              }}
            >
              Coffee created! Log your first brew to start tracking.
            </div>
          )}

          {/* Flavor Wheel - Average Traits */}
          {brews.length > 0 && (() => {
            const avgTraits = computeAverageTraits(brews);
            if (!avgTraits) return null;
            return (
              <div className="pokemon-textbox mb-sm" style={{ textAlign: "center" }}>
                <div style={{ fontWeight: "bold", marginBottom: "8px", fontSize: "10px" }}>
                  FLAVOR PROFILE
                </div>
                <div style={{ display: "flex", justifyContent: "center" }}>
                  <FlavorWheel traits={avgTraits} size={200} showLabels={true} />
                </div>
                <div style={{ fontSize: "8px", marginTop: "4px", opacity: 0.7 }}>
                  Average of {brews.length} brew{brews.length !== 1 ? "s" : ""}
                </div>
              </div>
            );
          })()}

          {/* Brew Metrics Chart */}
          {brews.length >= 2 && (
            <div className="mb-sm">
              <BrewChart brews={brews} />
            </div>
          )}

          {/* Brews List */}
          <div className="pokemon-textbox" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              BREWS ({brews.length})
            </div>
            {brews.length === 0 ? (
              <div style={{ fontSize: "9px", opacity: 0.7 }}>
                No brews yet. Log your first brew!
              </div>
            ) : (
              <div>
                {brews.map((brew, i) => {
                  const brewDaysOff = brew.days_off_roast;
                  const brewStale = brewDaysOff >= 0 && isStale(brewDaysOff);
                  return (
                    <div
                      key={brew.id}
                      style={{
                        padding: "4px 0",
                        borderBottom:
                          i < brews.length - 1 ? "1px solid #ccc" : "none",
                      }}
                    >
                      <div style={{ display: "flex", justifyContent: "space-between" }}>
                        <span>
                          Rating: {brew.rating}/10 | {brew.dripper || "No dripper"}
                          {brewDaysOff >= 0 && (
                            <span style={{ color: brewStale ? "#cc0000" : "inherit" }}>
                              {" "}| {brewDaysOff}d{brewStale && " ⚠"}
                            </span>
                          )}
                        </span>
                        <span>
                          {brew.end_time.minutes}:{String(brew.end_time.seconds).padStart(2, "0")}
                        </span>
                      </div>
                      {brew.tasting_notes.filter((n) => n).length > 0 && (
                        <div style={{ fontSize: "8px", opacity: 0.8 }}>
                          {brew.tasting_notes.filter((n) => n).join(", ")}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  const renderBrewForm = () => {
    if (!state.currentCoffee) {
      return (
        <div className="pokemon-screen">
          <div
            className="pokemon-frame"
          >
            <div className="pokemon-textbox">No coffee selected</div>
          </div>
        </div>
      );
    }

    return (
      <BrewForm
        coffeeId={state.currentCoffee.id}
        coffee={state.currentCoffee}
        formData={brewFormData}
        setFormData={setBrewFormData}
        onSubmit={handleBrewSubmit}
        onBack={() => {
          resetBrewForm();
          setState((prev) => ({ ...prev, view: "coffee-detail" }));
        }}
        error={state.error}
      />
    );
  };

  const renderCoffeeForm = () => (
    <CoffeeForm
      formData={coffeeFormData}
      setFormData={setCoffeeFormData}
      onSubmit={handleCoffeeSubmit}
      onBack={() => {
        resetCoffeeForm();
        setState((prev) => ({ ...prev, view: "home" }));
      }}
      error={state.error}
    />
  );

  const renderPokedex = () => {
    if (state.loading) {
      return (
        <div className="pokemon-screen">
          <div className="pokemon-frame">
            <LoadingSpinner variant="pokedex" message="Loading Pokedex..." />
          </div>
        </div>
      );
    }

    if (state.pokedexAll.length === 0) {
      return (
        <div className="pokemon-screen">
          <div className="pokemon-frame">
            <button
              className="pokemon-button mb-md"
              onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
            >
              Back
            </button>
            <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
              COFFEEDEX
            </h2>
            <EmptyState variant="no-pokemon" />
          </div>
        </div>
      );
    }

    // If filters returned no results
    if (state.pokedex.length === 0) {
      return (
        <div className="pokemon-screen">
          <div className="pokemon-frame">
            <button
              className="pokemon-button mb-md"
              onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
            >
              Back
            </button>
            <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
              COFFEEDEX
            </h2>
            <div className="pokemon-textbox mb-sm" style={{ fontSize: "10px", textAlign: "center" }}>
              No Pokemon match your filters.
            </div>
            <button
              className="pokemon-button"
              onClick={() => {
                const cleared = { roaster: "", origin: "", variety: "", roast_level: "", processing_method: "", pokemon_type: "" };
                setState((prev) => ({ ...prev, pokedexFilters: cleared, loading: true }));
                applyFiltersAndSort(state.pokedexAll, cleared, state.pokedexSort, state.coffees);
              }}
            >
              Clear Filters
            </button>
          </div>
        </div>
      );
    }

    const pokemon =
      state.currentPokemon || state.pokedex[state.currentPokedexIndex];
    const coffee = state.currentCoffee;
    const brews = state.currentBrews || [];
    const spriteUrl = `./pokemon-sprites/${String(pokemon.pokemon_id).padStart(
      3,
      "0"
    )}.png`;

    const hasPrev = state.currentPokedexIndex > 0;
    const hasNext = state.currentPokedexIndex < state.pokedex.length - 1;

    if (!coffee) {
      return (
        <div className="pokemon-screen">
          <div className="pokemon-frame">
            <div className="pokemon-textbox">Loading coffee details...</div>
          </div>
        </div>
      );
    }

    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{
            position: "relative",
            borderColor: getTypeColor(pokemon.pokemon_type),
            boxShadow: `inset -2px -2px 0px rgba(0,0,0,0.2), inset 2px 2px 0px rgba(255,255,255,0.5), 0 0 12px ${getTypeColor(pokemon.pokemon_type)}40`
          }}
        >
          <FormDecoSprites seed={`pokedex-${pokemon.pokemon_id}`} spin={true} />

          {/* Back + Sort */}
          <button
            className="pokemon-button mb-md"
            onClick={() =>
              setState((prev) => ({
                ...prev,
                view: "home",
                justCreatedPokemon: false,
              }))
            }
            style={{ position: "relative", zIndex: 1 }}
          >
            Back
          </button>

          {/* Sort + Filter Controls */}
          {(() => {
            const activeFilterCount = Object.values(state.pokedexFilters).filter((v) => v !== "").length;

            // Build filter options from pokedex coffees
            const pokedexCoffeeIds = new Set(state.pokedexAll.map((p) => p.coffee_id));
            const pokedexCoffees = state.coffees.filter((c) => pokedexCoffeeIds.has(c.id));
            const uniqueVals = (fn: (c: Coffee) => string) => [...new Set(pokedexCoffees.map(fn).filter(Boolean))].sort();

            const filterOptions = {
              roaster: uniqueVals((c) => c.roaster),
              origin: uniqueVals((c) => c.origin),
              variety: uniqueVals((c) => c.variety),
              roast_level: uniqueVals((c) => c.roast_level),
              processing_method: uniqueVals((c) => c.processing_method),
              pokemon_type: [...new Set(
                state.pokedexAll.flatMap((p) => p.pokemon_type.split("/").map((t) => t.trim())).filter(Boolean)
              )].sort(),
            };

            return (
              <div className="pokemon-textbox mb-sm" style={{ fontSize: "9px", padding: "4px" }}>
                <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                  <span style={{ fontWeight: "bold" }}>Sort:</span>
                  <select
                    className="pokemon-select"
                    value={state.pokedexSort}
                    onChange={(e) => handleSortChange(e.target.value as SortOption)}
                    style={{ flex: 1, padding: "4px", fontSize: "9px", border: "1px solid #000" }}
                  >
                    <option value="pokedex">Pokedex # (Low-High)</option>
                    <option value="date">Date Added (Newest)</option>
                    <option value="roast_date">Roast Date (Newest)</option>
                    <option value="name">Name (A-Z)</option>
                    <option value="confidence">Confidence (High-Low)</option>
                  </select>
                  <button
                    className="pokemon-button"
                    onClick={() => setState((prev) => ({ ...prev, pokedexFiltersOpen: !prev.pokedexFiltersOpen }))}
                    style={{ fontSize: "9px", padding: "2px 6px", whiteSpace: "nowrap" }}
                  >
                    Filter{activeFilterCount > 0 ? ` (${activeFilterCount})` : ""} {state.pokedexFiltersOpen ? "▲" : "▼"}
                  </button>
                  <span style={{ whiteSpace: "nowrap" }}>
                    {state.pokedex.length > 0 ? `${state.currentPokedexIndex + 1}/${state.pokedex.length}` : "0/0"}
                    {state.pokedexAll.length !== state.pokedex.length && (
                      <span style={{ opacity: 0.6 }}> of {state.pokedexAll.length}</span>
                    )}
                  </span>
                </div>

                {state.pokedexFiltersOpen && (
                  <div style={{ marginTop: "6px", display: "grid", gridTemplateColumns: "1fr 1fr", gap: "4px" }}>
                    {([
                      ["roaster", "Roaster"],
                      ["origin", "Origin"],
                      ["variety", "Variety"],
                      ["roast_level", "Roast"],
                      ["processing_method", "Process"],
                      ["pokemon_type", "Type"],
                    ] as [keyof typeof filterOptions, string][]).map(([key, label]) => (
                      <div key={key}>
                        <div style={{ fontSize: "8px", opacity: 0.7, marginBottom: "1px" }}>{label}</div>
                        <select
                          value={state.pokedexFilters[key as keyof AppState["pokedexFilters"]]}
                          onChange={(e) => handleFilterChange(key, e.target.value)}
                          style={{ width: "100%", padding: "2px", fontSize: "9px", border: "1px solid #000" }}
                        >
                          <option value="">All</option>
                          {filterOptions[key].map((v) => (
                            <option key={v} value={v}>{v}</option>
                          ))}
                        </select>
                      </div>
                    ))}
                    {activeFilterCount > 0 && (
                      <button
                        className="pokemon-button"
                        onClick={() => {
                          const cleared = { roaster: "", origin: "", variety: "", roast_level: "", processing_method: "", pokemon_type: "" };
                          setState((prev) => ({ ...prev, pokedexFilters: cleared, loading: true }));
                          applyFiltersAndSort(state.pokedexAll, cleared, state.pokedexSort, state.coffees);
                        }}
                        style={{ gridColumn: "1 / -1", fontSize: "9px", padding: "2px 6px" }}
                      >
                        Clear Filters
                      </button>
                    )}
                  </div>
                )}
              </div>
            );
          })()}

          {/* Prev/Next Navigation */}
          <div style={{ display: "flex", gap: "4px", marginBottom: "8px" }}>
            <button
              className="pokemon-button"
              onClick={() => navigatePokedex("prev")}
              disabled={!hasPrev}
              style={{ flex: "0 0 auto" }}
            >
              ←
            </button>
            <div style={{ flex: 1 }} />
            <button
              className="pokemon-button"
              onClick={() => navigatePokedex("next")}
              disabled={!hasNext}
              style={{ flex: "0 0 auto" }}
            >
              →
            </button>
          </div>

          {/* Pokemon Header */}
          <div className="pokemon-textbox mb-sm" style={{ textAlign: "center" }}>
            <img
              src={spriteUrl}
              alt={pokemon.pokemon_name}
              className="pokemon-sprite"
              style={{
                width: "96px",
                height: "96px",
                display: "block",
                margin: "0 auto",
              }}
              onError={(e) => {
                e.currentTarget.style.display = "none";
              }}
            />
            <div style={{ fontSize: "14px", fontWeight: "bold", marginTop: "4px" }}>
              {pokemon.pokemon_name.toUpperCase()}
            </div>
            <div style={{ marginTop: "4px" }}>
              <span
                className="pokemon-badge"
                style={{ backgroundColor: getTypeColor(pokemon.pokemon_type) }}
              >
                {pokemon.pokemon_type}
              </span>
            </div>
            <div style={{ fontSize: "9px", marginTop: "4px", opacity: 0.7 }}>
              Level {pokemon.level}
            </div>
          </div>

          {/* Fun Description */}
          {pokemon.llm_description && (
            <div
              className="pokemon-textbox mb-sm"
              style={{ fontSize: "10px", lineHeight: "1.5", fontStyle: "italic" }}
            >
              {pokemon.llm_description}
            </div>
          )}

          {/* Coffee Info */}
          <div className="pokemon-textbox mb-sm" style={{ fontSize: "10px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              {coffee.name.toUpperCase()}
            </div>
            <div><strong>Origin:</strong> {coffee.origin}</div>
            <div><strong>Roaster:</strong> {coffee.roaster}</div>
            {coffee.variety && <div><strong>Variety:</strong> {coffee.variety}</div>}
            <div><strong>Roast:</strong> {coffee.roast_level}</div>
            <div><strong>Process:</strong> {coffee.processing_method}</div>
          </div>

          {/* Flavor Wheel */}
          {brews.length > 0 && (() => {
            const avgTraits = computeAverageTraits(brews);
            if (!avgTraits) return null;
            return (
              <div className="pokemon-textbox mb-sm" style={{ textAlign: "center" }}>
                <div style={{ fontWeight: "bold", marginBottom: "8px", fontSize: "10px" }}>
                  FLAVOR PROFILE
                </div>
                <div style={{ display: "flex", justifyContent: "center" }}>
                  <FlavorWheel traits={avgTraits} size={200} showLabels={true} />
                </div>
                <div style={{ fontSize: "8px", marginTop: "4px", opacity: 0.7 }}>
                  Average of {brews.length} brew{brews.length !== 1 ? "s" : ""}
                </div>
              </div>
            );
          })()}

          {/* Brew Chart */}
          {brews.length >= 2 && (
            <div className="mb-sm">
              <BrewChart brews={brews} />
            </div>
          )}

          {/* Brew List */}
          <div className="pokemon-textbox mb-sm" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              BREWS ({brews.length})
            </div>
            {brews.length === 0 ? (
              <div style={{ fontSize: "9px", opacity: 0.7 }}>
                No brews yet.
              </div>
            ) : (
              <div>
                {brews.map((brew, i) => (
                  <div
                    key={brew.id}
                    style={{
                      padding: "4px 0",
                      borderBottom: i < brews.length - 1 ? "1px solid #ccc" : "none",
                    }}
                  >
                    <div style={{ display: "flex", justifyContent: "space-between" }}>
                      <span>
                        Rating: {brew.rating}/10 | {brew.dripper || "No dripper"}
                      </span>
                      <span>
                        {brew.end_time.minutes}:{String(brew.end_time.seconds).padStart(2, "0")}
                      </span>
                    </div>
                    {brew.tasting_notes.filter((n) => n).length > 0 && (
                      <div style={{ fontSize: "8px", opacity: 0.8 }}>
                        {brew.tasting_notes.filter((n) => n).join(", ")}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Log Brew Button */}
          <button
            className="pokemon-button"
            onClick={() => {
              resetBrewForm();
              setBrewFormData((prev) => ({
                ...prev,
                coffee_id: coffee.id,
              }));
              setState((prev) => ({
                ...prev,
                view: "brew-form",
                formStep: 1,
              }));
            }}
            style={{ width: "100%", padding: "10px", fontSize: "11px" }}
          >
            Log Brew
          </button>

          {state.justCreatedPokemon && (
            <button
              className="pokemon-button mt-sm"
              onClick={() => {
                setState((prev) => ({
                  ...prev,
                  view: "home",
                  justCreatedPokemon: false,
                }));
              }}
              style={{ width: "100%", padding: "12px", fontSize: "11px" }}
            >
              Return to Home
            </button>
          )}
        </div>
      </div>
    );
  };

  const renderBrewSuccess = () => {
    const progress = state.brewProgress;
    const coffee = state.currentCoffee;
    if (!progress || !coffee) {
      return (
        <div className="pokemon-screen centered">
          <div className="pokemon-frame">
            <button className="pokemon-button mb-md" onClick={() => setState((prev) => ({ ...prev, view: "home" }))}>
              Home
            </button>
            <div className="pokemon-textbox">{state.brewSuccessMessage || "Brew saved!"}</div>
          </div>
        </div>
      );
    }

    const progressPercent = Math.min((progress.count / progress.required) * 100, 100);
    const remaining = Math.max(progress.required - progress.count, 0);
    const canGenerate = progress.can_generate_pokemon && !progress.has_pokemon;

    return (
      <div className="pokemon-screen centered">
        <div className="pokemon-frame" style={{ textAlign: "center", justifyContent: "center" }}>
          <h2 className="pokemon-title" style={{ fontSize: "14px" }}>BREW LOGGED!</h2>

          <div className="pokemon-textbox mb-md" style={{ fontSize: "11px" }}>
            {state.brewSuccessMessage}
          </div>

          <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>{coffee.name.toUpperCase()}</div>

            {/* Show Pokemon sprite if exists */}
            {progress.has_pokemon && state.currentPokemon && (
              <div style={{ margin: "12px 0" }}>
                <img
                  src={`./pokemon-sprites/animated/${state.currentPokemon.pokemon_id}.gif`}
                  alt={state.currentPokemon.pokemon_name}
                  style={{ width: "64px", height: "64px", imageRendering: "pixelated" }}
                  onError={(e) => {
                    e.currentTarget.src = `./pokemon-sprites/${String(state.currentPokemon!.pokemon_id).padStart(3, "0")}.png`;
                  }}
                />
                <div style={{ fontSize: "9px", marginTop: "4px", fontWeight: "bold" }}>
                  {state.currentPokemon.pokemon_name} is happy!
                </div>
              </div>
            )}

            {/* Animated Progress Bar */}
            <div style={{ fontWeight: "bold", marginBottom: "4px", marginTop: "8px" }}>
              POKEMON PROGRESS
            </div>
            <div className="pokemon-stat-bar" style={{ height: "12px", marginBottom: "4px" }}>
              <div
                className={`pokemon-stat-fill ${progress.has_pokemon ? "high" : canGenerate ? "medium" : "low"}`}
                style={{
                  width: `${progressPercent}%`,
                  transition: "width 1.2s ease-out",
                }}
              />
            </div>
            <div style={{ fontSize: "9px" }}>
              {progress.count}/{progress.required} brews
            </div>

            {/* Status message */}
            {progress.has_pokemon && (
              <div style={{ fontSize: "10px", marginTop: "8px", color: "#00aa00", fontWeight: "bold" }}>
                Pokemon Captured!
              </div>
            )}
            {!progress.has_pokemon && !canGenerate && remaining > 0 && (
              <div style={{ fontSize: "10px", marginTop: "8px" }}>
                {remaining} more brew{remaining !== 1 ? "s" : ""} to go!
              </div>
            )}
            {canGenerate && (
              <div style={{ fontSize: "10px", marginTop: "8px", color: "#aa6600", fontWeight: "bold" }}>
                Ready to discover a Pokemon!
              </div>
            )}
          </div>

          <div style={{ display: "flex", gap: "8px", justifyContent: "center" }}>
            <button
              className="pokemon-button"
              onClick={() => loadCoffeeDetail(coffee.id).then(() => setState((prev) => ({ ...prev, view: "coffee-detail" })))}
            >
              Back to Coffee
            </button>
            {canGenerate && (
              <button
                className="pokemon-button"
                onClick={handleGeneratePokemon}
                style={{ background: "#ffcc00" }}
              >
                Generate Pokemon!
              </button>
            )}
          </div>
        </div>
      </div>
    );
  };

  const renderPokemonEncounter = () => {
    const pokemon = state.encounterPokemon;
    if (!pokemon) {
      return (
        <div className="pokemon-screen centered">
          <div className="pokemon-frame">
            <LoadingSpinner variant="default" message="Something went wrong..." />
          </div>
        </div>
      );
    }

    const typeColor = getTypeColor(pokemon.pokemon_type);
    const spriteUrl = `./pokemon-sprites/animated/${pokemon.pokemon_id}.gif`;
    const staticSpriteUrl = `./pokemon-sprites/${String(pokemon.pokemon_id).padStart(3, "0")}.png`;

    return (
      <div className="pokemon-screen centered" style={{ position: "relative" }}>
        {/* White flash overlay */}
        {state.encounterPhase === "flash" && (
          <div style={{
            position: "fixed",
            top: 0, left: 0, right: 0, bottom: 0,
            background: "white",
            zIndex: 100,
            animation: "encounter-flash 0.6s ease-out forwards",
          }} />
        )}

        <div className="pokemon-frame" style={{
          textAlign: "center",
          justifyContent: "center",
          borderColor: typeColor,
          boxShadow: `inset -2px -2px 0px rgba(0,0,0,0.2), inset 2px 2px 0px rgba(255,255,255,0.5), 0 0 16px ${typeColor}60`,
        }}>
          {/* "Wild X appeared!" text */}
          {(state.encounterPhase === "text" || state.encounterPhase === "sprite" || state.encounterPhase === "done") && (
            <div className="pokemon-textbox mb-md" style={{
              fontSize: "12px",
              fontWeight: "bold",
              animation: state.encounterPhase === "text" ? "encounter-text 0.8s ease-out" : undefined,
            }}>
              Wild {pokemon.pokemon_name.toUpperCase()} appeared!
            </div>
          )}

          {/* Pokemon sprite slide-in */}
          {(state.encounterPhase === "sprite" || state.encounterPhase === "done") && (
            <div style={{
              margin: "16px 0",
              animation: state.encounterPhase === "sprite" ? "encounter-slide-in 0.5s ease-out" : undefined,
            }}>
              <div className="pokemon-sprite-container" style={{
                display: "inline-block",
                padding: "16px",
                borderColor: typeColor,
              }}>
                <img
                  src={spriteUrl}
                  alt={pokemon.pokemon_name}
                  style={{ width: "96px", height: "96px", imageRendering: "pixelated" }}
                  onError={(e) => { e.currentTarget.src = staticSpriteUrl; }}
                />
              </div>
              <div style={{ marginTop: "8px" }}>
                <span style={{ fontSize: "14px", fontWeight: "bold" }}>{pokemon.pokemon_name}</span>
                <span className="pokemon-badge" style={{ backgroundColor: typeColor, marginLeft: "8px" }}>
                  {pokemon.pokemon_type}
                </span>
              </div>
              <div style={{ fontSize: "10px", marginTop: "4px" }}>Level {pokemon.level}</div>
            </div>
          )}

          {/* Done - show action buttons */}
          {state.encounterPhase === "done" && (
            <div style={{ marginTop: "16px" }}>
              <div className="pokemon-textbox mb-md" style={{ fontSize: "10px", lineHeight: "1.4" }}>
                {pokemon.llm_description}
              </div>
              <div style={{ display: "flex", gap: "8px", justifyContent: "center" }}>
                <button
                  className="pokemon-button"
                  onClick={() => {
                    loadPokedex();
                    setState((prev) => ({
                      ...prev,
                      view: "pokedex",
                      justCreatedPokemon: true,
                      encounterPokemon: null,
                    }));
                  }}
                >
                  View in Pokedex
                </button>
                <button
                  className="pokemon-button"
                  onClick={() => setState((prev) => ({
                    ...prev,
                    view: "home",
                    encounterPokemon: null,
                    justCreatedPokemon: false,
                  }))}
                >
                  Home
                </button>
              </div>
            </div>
          )}

          {/* Click/tap to advance through phases */}
          {state.encounterPhase !== "done" && (
            <div style={{ fontSize: "9px", opacity: 0.7, marginTop: "12px", animation: "arrow-blink 0.8s infinite" }}>
              Click to continue...
            </div>
          )}
        </div>

        {/* Click handler to advance phases */}
        {state.encounterPhase !== "done" && (
          <div
            style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, zIndex: state.encounterPhase === "flash" ? 101 : 1, cursor: "pointer" }}
            onClick={() => {
              const phases: Array<"flash" | "text" | "sprite" | "done"> = ["flash", "text", "sprite", "done"];
              const currentIdx = phases.indexOf(state.encounterPhase);
              if (currentIdx < phases.length - 1) {
                setState((prev) => ({ ...prev, encounterPhase: phases[currentIdx + 1] }));
              }
            }}
          />
        )}
      </div>
    );
  };

  const renderSettings = () => (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ position: "relative", justifyContent: "center" }}
      >
        <FormDecoSprites seed="settings" spin={true} />
        <button
          className="pokemon-button mb-md"
          onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
          style={{ position: "relative", zIndex: 1 }}
        >
          Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px", position: "relative", zIndex: 1 }}>
          SETTINGS
        </h2>

        <div className="pokemon-textbox mb-md">
          <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
            Color Theme
          </div>
          <div style={{ fontSize: "10px", marginBottom: "12px" }}>
            Select your Game Boy Color theme:
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
            <button
              className="pokemon-button"
              onClick={() =>
                setState((prev) => ({ ...prev, colorTheme: "blue" }))
              }
              style={{
                background: state.colorTheme === "blue" ? "#0066cc" : undefined,
                color: state.colorTheme === "blue" ? "white" : undefined,
              }}
            >
              Blue {state.colorTheme === "blue" ? "OK" : ""}
            </button>
            <button
              className="pokemon-button"
              onClick={() =>
                setState((prev) => ({ ...prev, colorTheme: "red" }))
              }
              style={{
                background: state.colorTheme === "red" ? "#cc0000" : undefined,
                color: state.colorTheme === "red" ? "white" : undefined,
              }}
            >
              Red {state.colorTheme === "red" ? "OK" : ""}
            </button>
            <button
              className="pokemon-button"
              onClick={() =>
                setState((prev) => ({ ...prev, colorTheme: "yellow" }))
              }
              style={{
                background:
                  state.colorTheme === "yellow" ? "#ccaa00" : undefined,
                color: state.colorTheme === "yellow" ? "white" : undefined,
              }}
            >
              Yellow {state.colorTheme === "yellow" ? "OK" : ""}
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  if (state.loading && state.view === "coffee-form") {
    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
        >
          <LoadingSpinner variant="save" message="Creating Coffee..." />
        </div>
      </div>
    );
  }

  if (state.loading && state.view === "brew-form") {
    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
        >
          <LoadingSpinner variant="brew" message="Saving Brew..." />
        </div>
      </div>
    );
  }

  return (
    <div data-theme={state.colorTheme}>
      <TitleBar />
      <div style={{ paddingTop: "32px" }}>
        {state.view === "start" && renderStart()}
        {state.view === "home" && renderHome()}
        {state.view === "coffee-form" && renderCoffeeForm()}
        {state.view === "coffee-list" && renderCoffeeList()}
        {state.view === "recent-brews" && renderRecentBrews()}
        {state.view === "coffee-detail" && renderCoffeeDetail()}
        {state.view === "brew-form" && renderBrewForm()}
        {state.view === "pokedex" && renderPokedex()}
        {state.view === "brew-success" && renderBrewSuccess()}
        {state.view === "pokemon-encounter" && renderPokemonEncounter()}
        {state.view === "settings" && renderSettings()}
        {state.view === "statistics" && (
          <Statistics
            onBack={() => setState((prev) => ({ ...prev, view: "home" }))}
          />
        )}
        {state.view === "special-items" && (
          <SpecialItems
            onBack={() => setState((prev) => ({ ...prev, view: "home" }))}
          />
        )}
        {state.view === "comparison" && (
          <CoffeeComparison
            onBack={() => setState((prev) => ({ ...prev, view: "home" }))}
          />
        )}
      </div>
    </div>
  );
};

export default App;
