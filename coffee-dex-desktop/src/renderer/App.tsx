import React, { useState, useEffect } from "react";
import { api } from "../services/api";
import {
  Coffee,
  CoffeePokemon,
  TastingTraits,
  Brew,
  BrewProgress,
} from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import CoffeeForm from "./CoffeeForm";
import BrewForm from "./BrewForm";
import Statistics from "./Statistics";
import SpecialItems from "./SpecialItems";
import TitleBar from "../components/TitleBar";

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
    | "brew-success";
  coffees: Coffee[];
  recentCoffees: Coffee[];
  currentCoffee: Coffee | null;
  currentPokemon: CoffeePokemon | null;
  currentBrews: Brew[];
  brewProgress: BrewProgress | null;
  pokedex: CoffeePokemon[];
  currentPokedexIndex: number;
  loading: boolean;
  error: string | null;
  backendConnected: boolean;
  formStep: number;
  pokedexPage: number;
  colorTheme: "red" | "blue" | "yellow";
  pokedexSort: "date" | "rating" | "name" | "confidence";
  justCreatedPokemon: boolean;
  promptFirstBrew: boolean; // After creating coffee, prompt to add first brew
}

const defaultBrewFormData = (): Partial<Brew> => ({
  coffee_id: "",
  tasting_notes: ["", "", "", "", ""],
  rating: 5,
  recipe: [],
  dripper: "",
  end_time: { minutes: 0, seconds: 0 },
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

const App: React.FC = () => {
  const [state, setState] = useState<AppState>({
    view: "start",
    coffees: [],
    recentCoffees: [],
    currentCoffee: null,
    currentPokemon: null,
    currentBrews: [],
    brewProgress: null,
    pokedex: [],
    currentPokedexIndex: 0,
    loading: false,
    error: null,
    backendConnected: false,
    formStep: 1,
    pokedexPage: 1,
    colorTheme: "blue",
    pokedexSort: "date",
    justCreatedPokemon: false,
    promptFirstBrew: false,
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
      let pokedex = await api.getPokedex();
      pokedex = sortPokedex(pokedex, state.pokedexSort);

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

  const sortPokedex = (
    pokedex: CoffeePokemon[],
    sortBy: "date" | "rating" | "name" | "confidence"
  ) => {
    const sorted = [...pokedex];
    switch (sortBy) {
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
      default:
        return sorted;
    }
  };

  const handleSortChange = async (
    sortBy: "date" | "rating" | "name" | "confidence"
  ) => {
    setState((prev) => ({ ...prev, pokedexSort: sortBy, loading: true }));
    const sortedPokedex = sortPokedex(state.pokedex, sortBy);

    if (sortedPokedex.length > 0) {
      try {
        const firstPokemon = sortedPokedex[0];
        const coffee = await api.getCoffee(firstPokemon.coffee_id);
        setState((prev) => ({
          ...prev,
          pokedex: sortedPokedex,
          currentPokemon: firstPokemon,
          currentCoffee: coffee,
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
      setState((prev) => ({ ...prev, pokedex: sortedPokedex, loading: false }));
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

      // Reload coffee detail to update brew count
      await loadCoffeeDetail(state.currentCoffee.id);
      resetBrewForm();
      setState((prev) => ({
        ...prev,
        loading: false,
        view: "coffee-detail",
        promptFirstBrew: false,
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
      setState((prev) => ({
        ...prev,
        currentPokemon: pokemon,
        view: "pokedex",
        loading: false,
        justCreatedPokemon: true,
      }));
      // Reload pokedex
      await loadPokedex();
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
            COFFEEDEX
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
        <h1 className="pokemon-title">COFFEEDEX</h1>
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
          Log 5 brews of a coffee to generate its Pokemon!
        </div>

        <div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "12px",
              maxWidth: "500px",
              margin: "24px auto 0",
            }}
          >
            <button
              className="pokemon-button"
              onClick={() => {
                resetCoffeeForm();
                setState((prev) => ({
                  ...prev,
                  view: "coffee-form",
                  formStep: 1,
                }));
              }}
              disabled={!state.backendConnected}
              style={{ fontSize: "11px", padding: "10px" }}
            >
              New Coffee
            </button>
            <button
              className="pokemon-button"
              onClick={async () => {
                await loadRecentCoffees();
                setState((prev) => ({
                  ...prev,
                  view: "coffee-list",
                }));
              }}
              disabled={!state.backendConnected}
              style={{ fontSize: "11px", padding: "10px" }}
            >
              Quick Brew
            </button>
            <button
              className="pokemon-button"
              onClick={async () => {
                await loadCoffees();
                setState((prev) => ({ ...prev, view: "coffee-list" }));
              }}
              disabled={!state.backendConnected}
              style={{ fontSize: "11px", padding: "10px" }}
            >
              View Coffees
            </button>
            <button
              className="pokemon-button"
              onClick={() => {
                loadPokedex();
                setState((prev) => ({ ...prev, view: "pokedex" }));
              }}
              disabled={!state.backendConnected}
              style={{ fontSize: "11px", padding: "10px" }}
            >
              Pokedex
            </button>
            <button
              className="pokemon-button"
              onClick={() =>
                setState((prev) => ({ ...prev, view: "statistics" }))
              }
              disabled={!state.backendConnected}
              style={{ fontSize: "11px", padding: "10px" }}
            >
              Statistics
            </button>
            <button
              className="pokemon-button"
              onClick={() =>
                setState((prev) => ({ ...prev, view: "special-items" }))
              }
              disabled={!state.backendConnected}
              style={{ fontSize: "11px", padding: "10px" }}
            >
              Brewers
            </button>
          </div>
        </div>
      </div>
    </div>
  );

  const renderCoffeeList = () => {
    const coffees =
      state.coffees.length > 0 ? state.coffees : state.recentCoffees;

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
            Back
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
            COFFEES
          </h2>

          {coffees.length === 0 ? (
            <div className="pokemon-textbox text-center">
              <div style={{ fontSize: "10px" }}>No coffees yet!</div>
              <div style={{ fontSize: "9px", marginTop: "8px" }}>
                Add your first coffee to get started.
              </div>
            </div>
          ) : (
            <div style={{ maxHeight: "400px", overflowY: "auto" }}>
              {coffees.map((coffee) => (
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
                  <div style={{ fontWeight: "bold", fontSize: "11px" }}>
                    {coffee.name}
                  </div>
                  <div style={{ fontSize: "9px" }}>
                    {coffee.origin} | {coffee.roaster}
                  </div>
                  <div style={{ fontSize: "8px", marginTop: "4px" }}>
                    {coffee.roast_level} | {coffee.processing_method}
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
            style={{ maxWidth: "600px", margin: "0 auto" }}
          >
            <div className="pokemon-textbox">Loading...</div>
          </div>
        </div>
      );
    }

    const coffee = state.currentCoffee;
    const progress = state.brewProgress;
    const brews = state.currentBrews;

    const progressPercent = Math.min(
      (progress.count / progress.required) * 100,
      100
    );
    const canGenerate = progress.can_generate_pokemon && !progress.has_pokemon;

    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button
            className="pokemon-button mb-md"
            onClick={() =>
              setState((prev) => ({
                ...prev,
                view: "coffee-list",
                promptFirstBrew: false,
              }))
            }
          >
            Back
          </button>

          <h2
            className="pokemon-title"
            style={{ fontSize: "14px", marginBottom: "8px" }}
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
              gap: "8px",
              marginBottom: "12px",
            }}
          >
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
              <div style={{ maxHeight: "150px", overflowY: "auto" }}>
                {brews.map((brew, i) => (
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
            style={{ maxWidth: "600px", margin: "0 auto" }}
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
        formStep={state.formStep}
        setFormStep={(step) =>
          setState((prev) => ({ ...prev, formStep: step }))
        }
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
      formStep={state.formStep}
      setFormStep={(step) => setState((prev) => ({ ...prev, formStep: step }))}
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
          <div
            className="pokemon-frame"
            style={{ maxWidth: "600px", margin: "0 auto" }}
          >
            <div className="pokemon-loading">Loading Pokedex</div>
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
              Back
            </button>
            <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
              COFFEEDEX
            </h2>
            <div className="pokemon-textbox text-center">
              <div style={{ fontSize: "10px" }}>No Pokemon yet!</div>
              <div style={{ fontSize: "8px", marginTop: "8px" }}>
                Log 5 brews of a coffee to generate your first Pokemon.
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
                setState((prev) => ({
                  ...prev,
                  view: "home",
                  pokedexPage: 1,
                  justCreatedPokemon: false,
                }))
              }
            >
              Back
            </button>

            <div className="mb-sm">
              <div
                className="pokemon-textbox mb-sm"
                style={{ fontSize: "9px", padding: "4px" }}
              >
                <div
                  style={{ display: "flex", alignItems: "center", gap: "8px" }}
                >
                  <span style={{ fontWeight: "bold" }}>Sort:</span>
                  <select
                    className="pokemon-select"
                    value={state.pokedexSort}
                    onChange={(e) =>
                      handleSortChange(
                        e.target.value as
                          | "date"
                          | "rating"
                          | "name"
                          | "confidence"
                      )
                    }
                    style={{
                      flex: 1,
                      padding: "4px",
                      fontSize: "9px",
                      border: "1px solid #000",
                    }}
                  >
                    <option value="date">Date (Newest)</option>
                    <option value="name">Name (A-Z)</option>
                    <option value="confidence">Confidence (High-Low)</option>
                  </select>
                </div>
              </div>
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
                <strong>Pokemon:</strong> {pokemon.pokemon_name}
              </div>
              <div>
                <strong>Level:</strong> {pokemon.level}
              </div>
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
            </div>

            <div className="pokemon-nav mt-md">
              <button
                className="pokemon-button"
                onClick={() => navigatePokedex("prev")}
                disabled={!hasPrev}
              >
                Prev
              </button>
              <button
                className="pokemon-button"
                onClick={() =>
                  setState((prev) => ({ ...prev, pokedexPage: 2 }))
                }
              >
                Analysis
              </button>
              <button
                className="pokemon-button"
                onClick={() => navigatePokedex("next")}
                disabled={!hasNext}
              >
                Next
              </button>
            </div>

            {state.justCreatedPokemon && (
              <button
                className="pokemon-button mt-md"
                onClick={() => {
                  setState((prev) => ({
                    ...prev,
                    view: "home",
                    pokedexPage: 1,
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
              setState((prev) => ({
                ...prev,
                view: "home",
                pokedexPage: 1,
                justCreatedPokemon: false,
              }))
            }
          >
            Back
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
                    - {tm.trait} → {tm.pokemon_stat}
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
              Prev
            </button>
            <button
              className="pokemon-button"
              onClick={() => setState((prev) => ({ ...prev, pokedexPage: 1 }))}
            >
              Details
            </button>
            <button
              className="pokemon-button"
              onClick={() => navigatePokedex("next")}
              disabled={!hasNext}
            >
              Next
            </button>
          </div>

          {state.justCreatedPokemon && (
            <button
              className="pokemon-button mt-md"
              onClick={() => {
                setState((prev) => ({
                  ...prev,
                  view: "home",
                  pokedexPage: 1,
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
          Back
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
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <div className="pokemon-loading">Creating Coffee</div>
        </div>
      </div>
    );
  }

  if (state.loading && state.view === "brew-form") {
    return (
      <div className="pokemon-screen">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <div className="pokemon-loading">Saving Brew</div>
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
        {state.view === "coffee-detail" && renderCoffeeDetail()}
        {state.view === "brew-form" && renderBrewForm()}
        {state.view === "pokedex" && renderPokedex()}
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
      </div>
    </div>
  );
};

export default App;
