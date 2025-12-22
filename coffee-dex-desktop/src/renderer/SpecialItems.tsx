import React, { useState, useEffect } from "react";
import { api } from "../services/api";

interface Recipe {
  id: string;
  name: string;
  steps: string[];
}

interface Brewer {
  id: string;
  name: string;
  pokeball_type: string;
  recipes: Recipe[];
  created_at: string;
}

interface SpecialItemsProps {
  onBack: () => void;
}

const POKEBALL_SPRITES: { [key: string]: string } = {
  "poke-ball": "left-poke-ball.png",
  "great-ball": "lagreat-ball.png",
  "ultra-ball": "laultra-ball.png",
  "fast-ball": "fast-ball.png",
};

const SpecialItems: React.FC<SpecialItemsProps> = ({ onBack }) => {
  const [brewers, setBrewers] = useState<Brewer[]>([]);
  const [selectedBrewer, setSelectedBrewer] = useState<Brewer | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newBrewerName, setNewBrewerName] = useState("");
  const [selectedPokeballType, setSelectedPokeballType] = useState("poke-ball");
  const [showAddRecipe, setShowAddRecipe] = useState(false);
  const [recipeSteps, setRecipeSteps] = useState<string[]>([""]);
  const [recipeName, setRecipeName] = useState("");

  useEffect(() => {
    loadBrewers();
  }, []);

  const loadBrewers = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getBrewers();
      setBrewers(data || []);
    } catch (err) {
      console.error("Failed to load brewers:", err);
      setError(`Failed to load brewers: ${err}`);
      setBrewers([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateBrewer = async () => {
    if (!newBrewerName.trim()) {
      setError("Brewer name cannot be empty");
      return;
    }

    try {
      await api.createBrewer(newBrewerName, selectedPokeballType);
      setNewBrewerName("");
      setShowCreateForm(false);
      await loadBrewers();
    } catch (err) {
      setError(`Failed to create brewer: ${err}`);
    }
  };

  const handleDeleteBrewer = async (brewerId: string) => {
    if (
      !confirm(
        "Are you sure you want to delete this brewer and all its recipes?"
      )
    ) {
      return;
    }

    try {
      await api.deleteBrewer(brewerId);
      await loadBrewers();
      setSelectedBrewer(null);
    } catch (err) {
      setError(`Failed to delete brewer: ${err}`);
    }
  };

  const handleRemoveRecipe = async (brewerId: string, recipeId: string) => {
    if (!confirm("Remove this recipe from the brewer?")) {
      return;
    }

    try {
      await api.removeStandaloneRecipe(brewerId, recipeId);
      await loadBrewers();
      // Update selected brewer
      if (brewers && brewers.length > 0) {
        const updated = brewers.find((b) => b.id === brewerId);
        if (updated) {
          setSelectedBrewer(updated);
        }
      }
    } catch (err) {
      setError(`Failed to remove recipe: ${err}`);
    }
  };

  const handleAddRecipeStep = () => {
    setRecipeSteps([...recipeSteps, ""]);
  };

  const handleRemoveRecipeStep = (index: number) => {
    if (recipeSteps.length > 1) {
      setRecipeSteps(recipeSteps.filter((_, i) => i !== index));
    }
  };

  const handleUpdateRecipeStep = (index: number, value: string) => {
    const updated = [...recipeSteps];
    updated[index] = value;
    setRecipeSteps(updated);
  };

  const handleFinishRecipe = async () => {
    if (!selectedBrewer) {
      setError("No brewer selected");
      return;
    }

    if (!recipeName.trim()) {
      setError("Please enter a recipe name");
      return;
    }

    // Filter out empty steps
    const validSteps = recipeSteps.filter((step) => step.trim() !== "");

    if (validSteps.length === 0) {
      setError("Please add at least one recipe step");
      return;
    }

    try {
      // Add standalone recipe to brewer
      await api.addStandaloneRecipe(selectedBrewer.id, recipeName, validSteps);

      // Reset form
      setRecipeSteps([""]);
      setRecipeName("");
      setShowAddRecipe(false);

      // Reload brewers
      await loadBrewers();

      // Update selected brewer
      const updated = brewers.find((b) => b.id === selectedBrewer.id);
      if (updated) {
        setSelectedBrewer(updated);
      }
    } catch (err) {
      setError(`Failed to add recipe: ${err}`);
    }
  };

  if (loading) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <div className="pokemon-loading">Loading Special Items</div>
        </div>
      </div>
    );
  }

  // Create Brewer Form
  if (showCreateForm) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "280px", margin: "0 auto" }}
        >
          <button
            className="pokemon-button mb-md"
            onClick={() => setShowCreateForm(false)}
          >
            ← Back
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "12px" }}>
            CREATE BREWER
          </h2>

          {error && (
            <div className="pokemon-textbox mb-md" style={{ color: "#cc0000" }}>
              {error}
            </div>
          )}

          <div className="pokemon-textbox mb-md">
            <div
              style={{
                fontWeight: "bold",
                marginBottom: "8px",
                fontSize: "8px",
              }}
            >
              Brewer Name
            </div>
            <input
              type="text"
              className="pokemon-input"
              value={newBrewerName}
              onChange={(e) => setNewBrewerName(e.target.value)}
              placeholder="Enter brewer name"
            />
          </div>

          <div className="pokemon-textbox mb-md">
            <div
              style={{
                fontWeight: "bold",
                marginBottom: "8px",
                fontSize: "8px",
              }}
            >
              Pokeball Type
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr",
                gap: "8px",
              }}
            >
              {Object.entries(POKEBALL_SPRITES).map(([type, sprite]) => (
                <button
                  key={type}
                  className="pokemon-button"
                  onClick={() => setSelectedPokeballType(type)}
                  style={{
                    background:
                      selectedPokeballType === type ? "#0066cc" : undefined,
                    color: selectedPokeballType === type ? "white" : undefined,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "flex-start",
                    gap: "8px",
                    fontSize: "8px",
                  }}
                >
                  <img
                    src={`./pokemon-sprites/${sprite}`}
                    alt={type}
                    style={{ width: "20px", height: "20px" }}
                    onError={(e) => {
                      const img = e.target as HTMLImageElement;
                      if (!img.dataset.errorHandled) {
                        img.dataset.errorHandled = "true";
                        img.src = "./pokemon-sprites/left-poke-ball.png";
                      }
                    }}
                  />
                  {type}
                </button>
              ))}
            </div>
          </div>

          <button className="pokemon-button" onClick={handleCreateBrewer}>
            Create Brewer
          </button>
        </div>
      </div>
    );
  }

  // Add Recipe View
  if (showAddRecipe && selectedBrewer) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "280px", margin: "0 auto" }}
        >
          <button
            className="pokemon-button mb-md"
            onClick={() => {
              setShowAddRecipe(false);
              setRecipeSteps([""]);
              setRecipeName("");
              setError(null);
            }}
          >
            ← Back
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "12px" }}>
            ADD RECIPE
          </h2>

          {error && (
            <div
              className="pokemon-textbox mb-md"
              style={{ color: "##cc0000", fontSize: "8px" }}
            >
              {error}
            </div>
          )}

          <div className="pokemon-textbox mb-md" style={{ fontSize: "8px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              Recipe Name
            </div>
            <input
              type="text"
              className="pokemon-input mb-md"
              value={recipeName}
              onChange={(e) => setRecipeName(e.target.value)}
              placeholder="e.g., V60 Pour Over"
            />

            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              Brewing Steps
            </div>
            {recipeSteps.map((step, index) => (
              <div
                key={index}
                style={{
                  marginBottom: "8px",
                  display: "flex",
                  gap: "8px",
                  alignItems: "center",
                }}
              >
                <div style={{ fontWeight: "bold", minWidth: "20px" }}>
                  {index + 1}.
                </div>
                <input
                  type="text"
                  className="pokemon-input"
                  value={step}
                  onChange={(e) =>
                    handleUpdateRecipeStep(index, e.target.value)
                  }
                  placeholder={`Step ${index + 1}`}
                  style={{
                    flex: 1,
                  }}
                />
                {recipeSteps.length > 1 && (
                  <button
                    className="pokemon-button"
                    onClick={() => handleRemoveRecipeStep(index)}
                    style={{
                      fontSize: "10px",
                      padding: "4px 8px",
                      background: "#cc0000",
                      color: "white",
                    }}
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
            <button
              className="pokemon-button mt-sm"
              onClick={handleAddRecipeStep}
              style={{ fontSize: "8px" }}
            >
              + Add Step
            </button>
          </div>

          <button
            className="pokemon-button"
            onClick={handleFinishRecipe}
            style={{ fontSize: "10px" }}
          >
            Save Recipe
          </button>
        </div>
      </div>
    );
  }

  // Brewer Detail View
  if (selectedBrewer) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button
            className="pokemon-button mb-md"
            onClick={() => setSelectedBrewer(null)}
          >
            ← Back
          </button>

          <div style={{ textAlign: "center", marginBottom: "16px" }}>
            <img
              src={`./pokemon-sprites/${
                POKEBALL_SPRITES[selectedBrewer.pokeball_type] ||
                "left-poke-ball.png"
              }`}
              alt={selectedBrewer.pokeball_type}
              style={{ width: "96px", height: "96px" }}
              onError={(e) => {
                const img = e.target as HTMLImageElement;
                if (!img.dataset.errorHandled) {
                  img.dataset.errorHandled = "true";
                  img.src = "./pokemon-sprites/left-poke-ball.png";
                }
              }}
            />
            <h2
              className="pokemon-title"
              style={{ fontSize: "14px", marginTop: "8px" }}
            >
              {selectedBrewer.name.toUpperCase()}
            </h2>
          </div>

          {error && (
            <div className="pokemon-textbox mb-md" style={{ color: "#cc0000" }}>
              {error}
            </div>
          )}

          <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              RECIPES (
              {selectedBrewer.recipes ? selectedBrewer.recipes.length : 0}/4)
            </div>
            {!selectedBrewer.recipes || selectedBrewer.recipes.length === 0 ? (
              <div style={{ textAlign: "center", opacity: 0.6 }}>
                No recipes yet
              </div>
            ) : (
              selectedBrewer.recipes.map((recipe) => (
                <div
                  key={recipe.id}
                  style={{
                    marginBottom: "8px",
                    padding: "8px",
                    background: "#f0f0f0",
                    borderRadius: "4px",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: "bold" }}>{recipe.name}</div>
                      <div style={{ fontSize: "8px" }}>
                        {recipe.steps.length} steps
                      </div>
                    </div>
                    <button
                      className="pokemon-button"
                      onClick={() =>
                        handleRemoveRecipe(selectedBrewer.id, recipe.id)
                      }
                      style={{
                        fontSize: "10px",
                        padding: "4px 8px",
                        background: "#cc0000",
                        color: "white",
                      }}
                    >
                      🗑️ Remove
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>

          {(!selectedBrewer.recipes || selectedBrewer.recipes.length < 4) && (
            <button
              className="pokemon-button mb-sm"
              onClick={() => setShowAddRecipe(true)}
            >
              + Add Recipe
            </button>
          )}

          <button
            className="pokemon-button"
            onClick={() => handleDeleteBrewer(selectedBrewer.id)}
            style={{ background: "#cc0000", color: "white" }}
          >
            Delete Brewer
          </button>
        </div>
      </div>
    );
  }

  // Gallery View
  return (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ maxWidth: "600px", margin: "0 auto" }}
      >
        <button className="pokemon-button mb-md" onClick={onBack}>
          ← Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
          SPECIAL ITEMS
        </h2>

        {error && (
          <div className="pokemon-textbox mb-md" style={{ color: "#cc0000" }}>
            {error}
          </div>
        )}

        <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
          Collect your favorite brews! Each pokeball represents a brewer with up
          to 4 signature recipes.
        </div>

        {!brewers || brewers.length === 0 ? (
          <div
            className="pokemon-textbox mb-md"
            style={{ textAlign: "center" }}
          >
            <div style={{ fontSize: "24px", marginBottom: "8px" }}>⚡</div>
            <div>No brewers yet!</div>
            <div style={{ fontSize: "8px", marginTop: "8px" }}>
              Create your first brewer to start collecting recipes.
            </div>
          </div>
        ) : (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "16px",
              marginBottom: "16px",
            }}
          >
            {brewers &&
              brewers.map((brewer) => (
                <button
                  key={brewer.id}
                  className="pokemon-button"
                  onClick={() => setSelectedBrewer(brewer)}
                  style={{
                    padding: "16px",
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    gap: "8px",
                  }}
                >
                  <img
                    src={`./pokemon-sprites/${
                      POKEBALL_SPRITES[brewer.pokeball_type] ||
                      "left-poke-ball.png"
                    }`}
                    alt={brewer.pokeball_type}
                    style={{ width: "64px", height: "64px" }}
                    onError={(e) => {
                      const img = e.target as HTMLImageElement;
                      if (!img.dataset.errorHandled) {
                        console.error(
                          "Failed to load pokeball sprite:",
                          brewer.pokeball_type
                        );
                        img.dataset.errorHandled = "true";
                        img.src = "./pokemon-sprites/left-poke-ball.png";
                      }
                    }}
                  />
                  <div style={{ fontSize: "10px", fontWeight: "bold" }}>
                    {brewer.name}
                  </div>
                  <div style={{ fontSize: "8px", opacity: 0.7 }}>
                    {brewer.recipes ? brewer.recipes.length : 0}/4 recipes
                  </div>
                </button>
              ))}
          </div>
        )}

        {brewers && brewers.length < 4 && (
          <button
            className="pokemon-button"
            onClick={() => setShowCreateForm(true)}
          >
            + Create Brewer
          </button>
        )}

        {brewers && brewers.length >= 4 && (
          <div
            className="pokemon-textbox"
            style={{ fontSize: "8px", textAlign: "center" }}
          >
            Maximum of 4 brewers reached!
          </div>
        )}
      </div>
    </div>
  );
};

export default SpecialItems;
