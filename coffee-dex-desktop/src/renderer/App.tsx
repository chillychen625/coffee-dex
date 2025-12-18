import React, { useState, useEffect } from "react";
import { api } from "../services/api";
import { Coffee, CoffeePokemon, TastingTraits } from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import CoffeeForm from "./CoffeeForm";

interface AppState {
  view: "start" | "home" | "coffee-form" | "pokedex" | "settings";
  coffees: Coffee[];
  recentCoffees: Coffee[];
  currentCoffee: Coffee | null;
  currentPokemon: CoffeePokemon | null;
  pokedex: CoffeePokemon[];
  currentPokedexIndex: number; // Current position in pokedex array
  loading: boolean;
  error: string | null;
  backendConnected: boolean;
  formStep: number; // 1: Basic Info, 2: Roast/Process, 3: Tasting Notes, 4: Tasting Traits 1, 5: Tasting Traits 2, 6: Recipe/Timing
  pokedexPage: number; // 1: Coffee Details, 2: LLM Analysis
  colorTheme: "red" | "blue" | "yellow"; // Game Boy Color theme
  isQuickBrew: boolean; // Whether we're in quick brew mode (subsequent brew of same coffee)
}

const App: React.FC = () => {
  const [state, setState] = useState<AppState>({
    view: "start",
    coffees: [],
    recentCoffees: [],
    currentCoffee: null,
    currentPokemon: null,
    pokedex: [],
    currentPokedexIndex: 0,
    loading: false,
    error: null,
    backendConnected: false,
    formStep: 1,
    pokedexPage: 1,
    colorTheme: "blue",
    isQuickBrew: false,
  });

  const [formData, setFormData] = useState<Partial<Coffee>>({
    name: "",
    origin: "",
    roaster: "",
    roast_level: "medium",
    processing_method: "washed",
    tasting_notes: ["", "", "", "", ""],
    rating: 5,
    recipe: [],
    dripper: "",
    end_time: {
      minutes: 0,
      seconds: 0,
    },
    tasting_traits: {
      berry_intensity: 5,
      stonefruit_intensity: 5,
      roast_intensity: 5,
      citrus_fruits_intensity: 5,
      bitterness: 5,
      florality: 5,
      spice: 5,
      sweetness: 5,
      aromatic_intensity: 5,
      savory: 5,
      body: 5,
      cleanliness: 5,
    } as TastingTraits,
  });

  // Check backend connection on mount
  useEffect(() => {
    checkBackend();
  }, []);

  const checkBackend = async () => {
    const connected = await api.healthCheck();
    setState((prev) => ({ ...prev, backendConnected: connected }));
    if (!connected) {
      setState((prev) => ({
        ...prev,
        error: "Backend not connected. Please start the server.",
      }));
    }
  };

  const loadPokedex = async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const pokedex = await api.getPokedex();
      // Fetch the first coffee's details if available
      if (pokedex.length > 0) {
        const firstPokemon = pokedex[0];
        const coffee = await api.getCoffee(firstPokemon.coffee_id);
        setState((prev) => ({
          ...prev,
          pokedex,
          currentPokemon: firstPokemon,
          currentCoffee: coffee,
          currentPokedexIndex: 0,
          loading: false,
        }));
      } else {
        setState((prev) => ({ ...prev, pokedex, loading: false }));
      }
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to load Pokedex: ${error}`,
        loading: false,
      }));
    }
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
        const coffee = await api.getCoffee(pokemon.coffee_id);
        setState((prev) => ({
          ...prev,
          currentPokemon: pokemon,
          currentCoffee: coffee,
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

  const handleCoffeeSubmit = async (coffee: Partial<Coffee>) => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      if (state.isQuickBrew) {
        // Quick brew: just save the entry, no Pokemon generation
        const newCoffee = await api.createBrewEntry(coffee);
        setState((prev) => ({
          ...prev,
          currentCoffee: newCoffee,
          loading: false,
          view: "home",
          isQuickBrew: false,
        }));
      } else {
        // Full new coffee: generate Pokemon
        const newCoffee = await api.createCoffee(coffee);
        setState((prev) => ({
          ...prev,
          currentCoffee: newCoffee,
          loading: false,
        }));
        await handleGeneratePokemon(newCoffee.id);
      }
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to create coffee: ${error}`,
        loading: false,
      }));
    }
  };

  const handleGeneratePokemon = async (coffeeId: string) => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const pokemon = await api.generatePokemon(coffeeId);
      setState((prev) => ({
        ...prev,
        currentPokemon: pokemon,
        view: "pokedex",
        loading: false,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        error: `Failed to generate Pokemon: ${error}`,
        loading: false,
      }));
    }
  };

  const renderStart = () => (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ maxWidth: "600px", margin: "0 auto" }}
      >
        <div style={{ textAlign: "center" }}>
          <h1
            className="pokemon-title"
            style={{ fontSize: "24px", marginBottom: "60px" }}
          >
            ☕ COFFEEDEX
          </h1>
          <button
            className="pokemon-button"
            onClick={() => {
              checkBackend();
              setState((prev) => ({ ...prev, view: "home" }));
            }}
            style={{ fontSize: "14px", padding: "12px 24px" }}
          >
            Press Start
          </button>
        </div>
      </div>
    </div>
  );

  const renderHome = () => (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ maxWidth: "600px", margin: "0 auto" }}
      >
        <h1 className="pokemon-title">☕ COFFEEDEX</h1>
        <p className="pokemon-subtitle">Gotta Brew 'Em All!</p>

        {!state.backendConnected && (
          <div
            className="pokemon-textbox"
            style={{ background: "#ffcccc", borderColor: "#cc0000" }}
          >
            <div style={{ fontSize: "10px" }}>
              Backend not connected!
              <br />
              Start server: go run main.go -storage=mysql
            </div>
          </div>
        )}

        <div className="pokemon-textbox">
          Transform your coffee tasting notes into Pokemon!
        </div>

        <div>
          <div
            style={{
              display: "flex",
              gap: "16px",
              justifyContent: "center",
              marginTop: "24px",
            }}
          >
            <button
              className="pokemon-button"
              onClick={() => {
                setState((prev) => ({
                  ...prev,
                  view: "coffee-form",
                  isQuickBrew: false,
                }));
              }}
              disabled={!state.backendConnected}
            >
              New Coffee
            </button>
            <button
              className="pokemon-button"
              onClick={async () => {
                await loadRecentCoffees();
                setState((prev) => ({
                  ...prev,
                  view: "coffee-form",
                  isQuickBrew: true,
                  formStep: 1,
                }));
              }}
              disabled={!state.backendConnected}
            >
              Quick Brew
            </button>
            <button
              className="pokemon-button"
              onClick={() => {
                loadPokedex();
                setState((prev) => ({ ...prev, view: "pokedex" }));
              }}
              disabled={!state.backendConnected}
            >
              View Pokedex
            </button>
          </div>
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              marginTop: "16px",
            }}
          >
            <button
              className="pokemon-button"
              onClick={() =>
                setState((prev) => ({ ...prev, view: "settings" }))
              }
              disabled={!state.backendConnected}
            >
              Settings
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  const renderSettings = () => (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ maxWidth: "600px", margin: "0 auto" }}
      >
        <button
          className="pokemon-button mb-md"
          onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
        >
          ← Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
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
              Blue {state.colorTheme === "blue" ? "✓" : ""}
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
              Red {state.colorTheme === "red" ? "✓" : ""}
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
              Yellow {state.colorTheme === "yellow" ? "✓" : ""}
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  const renderCoffeeForm = () => {
    const handleSubmit = () => {
      if (!formData.name || !formData.origin) {
        setState((prev) => ({
          ...prev,
          error: "Please fill in required fields (name, origin)",
        }));
        return;
      }
      handleCoffeeSubmit(formData);
    };

    return (
      <CoffeeForm
        formData={formData}
        setFormData={setFormData}
        formStep={state.formStep}
        setFormStep={(step) =>
          setState((prev) => ({ ...prev, formStep: step }))
        }
        onSubmit={handleSubmit}
        onBack={() =>
          setState((prev) => ({
            ...prev,
            view: "home",
            formStep: 1,
            isQuickBrew: false,
          }))
        }
        error={state.error}
        isQuickBrew={state.isQuickBrew}
        recentCoffees={state.recentCoffees}
      />
    );
  };

  const renderPokedex = () => {
    if (state.loading) {
      return (
        <div className="pokemon-screen">
          <div
            className="pokemon-frame"
            style={{ maxWidth: "600px", margin: "0 auto" }}
          >
            <div className="pokemon-loading">Generating Pokemon</div>
          </div>
        </div>
      );
    }

    if (!state.currentPokemon && state.pokedex.length === 0) {
      return (
        <div className="pokemon-screen">
          <div
            className="pokemon-frame"
            style={{ maxWidth: "600px", margin: "0 auto" }}
          >
            <button
              className="pokemon-button mb-md"
              onClick={() => setState((prev) => ({ ...prev, view: "home" }))}
            >
              ← Back
            </button>
            <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
              COFFEEDEX
            </h2>
            <div className="pokemon-textbox text-center">
              <div style={{ fontSize: "10px" }}>📝</div>
              <div>No Coffee yet!</div>
              <div style={{ fontSize: "8px", marginTop: "8px" }}>
                Create a coffee to generate your first entry.
              </div>
            </div>
          </div>
        </div>
      );
    }

    const pokemon =
      state.currentPokemon || state.pokedex[state.currentPokedexIndex];
    const coffee = state.currentCoffee;
    const spriteUrl = `./pokemon-sprites/${String(pokemon.pokemon_id).padStart(
      3,
      "0"
    )}.png`;

    const hasPrev = state.currentPokedexIndex > 0;
    const hasNext = state.currentPokedexIndex < state.pokedex.length - 1;

    if (!coffee) {
      return (
        <div className="pokemon-screen">
          <div
            className="pokemon-frame"
            style={{ maxWidth: "600px", margin: "0 auto" }}
          >
            <div className="pokemon-textbox">Loading coffee details...</div>
          </div>
        </div>
      );
    }

    const confidencePercent = pokemon.mapping_confidence * 100;
    const hpClass =
      confidencePercent > 70
        ? "high"
        : confidencePercent > 40
        ? "medium"
        : "low";

    // Page 1: Coffee Details
    if (state.pokedexPage === 1) {
      return (
        <div className="pokemon-screen">
          <div
            className="pokemon-frame"
            style={{ maxWidth: "600px", margin: "0 auto" }}
          >
            <button
              className="pokemon-button mb-md"
              onClick={() =>
                setState((prev) => ({ ...prev, view: "home", pokedexPage: 1 }))
              }
            >
              ← Back
            </button>

            <div className="mb-sm">
              <div
                className="pokemon-textbox"
                style={{
                  fontSize: "9px",
                  textAlign: "center",
                  padding: "4px",
                  marginBottom: "8px",
                }}
              >
                Entry {state.currentPokedexIndex + 1} of {state.pokedex.length}
              </div>

              <div
                className="pokemon-sprite-container"
                style={{ textAlign: "center", padding: "4px 0" }}
              >
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
                <div
                  style={{
                    fontSize: "12px",
                    fontWeight: "bold",
                    marginTop: "4px",
                  }}
                >
                  {coffee.name.toUpperCase()}
                </div>
              </div>
            </div>

            <div className="pokemon-textbox mb-sm" style={{ fontSize: "10px" }}>
              <div>
                <strong>Origin:</strong> {coffee.origin}
              </div>
              <div>
                <strong>Roaster:</strong> {coffee.roaster}
              </div>
              <div>
                <strong>Roast:</strong> {coffee.roast_level}
              </div>
              <div>
                <strong>Process:</strong> {coffee.processing_method}
              </div>
              {coffee.dripper && (
                <div>
                  <strong>Brewer:</strong> {coffee.dripper}
                </div>
              )}
            </div>

            <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
              <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
                TASTING NOTES:
              </div>
              {coffee.tasting_notes
                .filter((n) => n)
                .map((note, i) => (
                  <div key={i}>▸ {note}</div>
                ))}
            </div>

            <div className="pokemon-textbox" style={{ fontSize: "8px" }}>
              <div
                style={{
                  fontWeight: "bold",
                  marginBottom: "4px",
                  textAlign: "center",
                }}
              >
                FLAVOR PROFILE
              </div>
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "4px",
                }}
              >
                {Object.entries(coffee.tasting_traits).map(([key, value]) => (
                  <div
                    key={key}
                    style={{
                      marginBottom: "2px",
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <div
                      style={{ fontSize: "7px", textTransform: "capitalize" }}
                    >
                      {key.replace(/_/g, " ")}
                    </div>
                    <div style={{ fontSize: "8px", fontWeight: "bold" }}>
                      {value}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="pokemon-nav mt-md">
              <button
                className="pokemon-button"
                onClick={() => navigatePokedex("prev")}
                disabled={!hasPrev}
              >
                ← Prev
              </button>
              <button
                className="pokemon-button"
                onClick={() =>
                  setState((prev) => ({ ...prev, pokedexPage: 2 }))
                }
              >
                Analysis →
              </button>
              <button
                className="pokemon-button"
                onClick={() => navigatePokedex("next")}
                disabled={!hasNext}
              >
                Next →
              </button>
            </div>
          </div>
        </div>
      );
    }

    // Page 2: LLM Analysis
    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button
            className="pokemon-button mb-md"
            onClick={() =>
              setState((prev) => ({ ...prev, view: "home", pokedexPage: 1 }))
            }
          >
            ← Back
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
            ANALYSIS
          </h2>

          <div
            className="pokemon-textbox mb-md"
            style={{ textAlign: "center" }}
          >
            <div style={{ fontSize: "12px", fontWeight: "bold" }}>
              {pokemon.pokemon_name.toUpperCase()}
            </div>
            <div style={{ fontSize: "10px", marginTop: "4px" }}>
              Level {pokemon.level}
            </div>
          </div>

          <div
            className="pokemon-textbox mb-md"
            style={{ fontSize: "10px", lineHeight: "1.4" }}
          >
            {pokemon.llm_description}
          </div>

          <div className="pokemon-form-group mb-md">
            <div className="pokemon-form-label">Mapping Confidence</div>
            <div className="pokemon-stat-row">
              <div className="pokemon-stat-bar" style={{ flex: 1 }}>
                <div
                  className={`pokemon-stat-fill ${hpClass}`}
                  style={{ width: `${confidencePercent}%` }}
                ></div>
              </div>
              <div className="pokemon-stat-value">
                {Math.round(confidencePercent)}%
              </div>
            </div>
          </div>

          {pokemon.trait_mapping && pokemon.trait_mapping.length > 0 && (
            <div className="pokemon-textbox" style={{ fontSize: "8px" }}>
              <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
                TRAIT MAPPING:
              </div>
              {pokemon.trait_mapping.slice(0, 5).map((tm, i) => (
                <div key={i} style={{ marginBottom: "4px" }}>
                  <div>
                    ▸ {tm.trait} → {tm.pokemon_stat}
                  </div>
                  <div
                    style={{ fontSize: "7px", marginLeft: "8px", opacity: 0.8 }}
                  >
                    {tm.reasoning}
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="pokemon-nav mt-md">
            <button
              className="pokemon-button"
              onClick={() => navigatePokedex("prev")}
              disabled={!hasPrev}
            >
              ← Prev
            </button>
            <button
              className="pokemon-button"
              onClick={() => setState((prev) => ({ ...prev, pokedexPage: 1 }))}
            >
              ← Details
            </button>
            <button
              className="pokemon-button"
              onClick={() => navigatePokedex("next")}
              disabled={!hasNext}
            >
              Next →
            </button>
          </div>
        </div>
      </div>
    );
  };

  if (state.loading && state.view === "coffee-form") {
    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <div className="pokemon-loading">Generating Pokemon</div>
        </div>
      </div>
    );
  }

  return (
    <div data-theme={state.colorTheme}>
      {state.view === "start" && renderStart()}
      {state.view === "home" && renderHome()}
      {state.view === "coffee-form" && renderCoffeeForm()}
      {state.view === "pokedex" && renderPokedex()}
      {state.view === "settings" && renderSettings()}
    </div>
  );
};

export default App;
