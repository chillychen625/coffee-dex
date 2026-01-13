import React, { useState, useEffect } from "react";
import { Coffee } from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import { api } from "../services/api";

interface CoffeeFormProps {
  formData: Partial<Coffee>;
  setFormData: (data: Partial<Coffee>) => void;
  formStep: number;
  setFormStep: (step: number) => void;
  onSubmit: () => void;
  onBack: () => void;
  error: string | null;
}

const CoffeeForm: React.FC<CoffeeFormProps> = ({
  formData,
  setFormData,
  formStep,
  setFormStep,
  onSubmit,
  onBack,
  error,
}) => {
  const [uniqueOrigins, setUniqueOrigins] = useState<string[]>([]);
  const [uniqueRoasters, setUniqueRoasters] = useState<string[]>([]);

  useEffect(() => {
    loadAllCoffeesForAutocomplete();
  }, []);

  const loadAllCoffeesForAutocomplete = async () => {
    try {
      const coffees = await api.getCoffees();

      // Extract unique origins
      const origins = new Set<string>();
      coffees.forEach((c) => {
        if (c.origin) origins.add(c.origin);
      });
      setUniqueOrigins(Array.from(origins).sort());

      // Extract unique roasters
      const roasters = new Set<string>();
      coffees.forEach((c) => {
        if (c.roaster) roasters.add(c.roaster);
      });
      setUniqueRoasters(Array.from(roasters).sort());
    } catch (err) {
      console.error("Failed to load coffees for autocomplete:", err);
    }
  };

  // Step 1: Basic coffee info (name, origin, roaster, variety)
  const renderStep1 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">BASIC INFO (1/2)</div>
      <input
        type="text"
        className="pokemon-input mb-sm"
        placeholder="Coffee Name *"
        value={formData.name || ""}
        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
      />
      <input
        type="text"
        className="pokemon-input mb-sm"
        placeholder="Origin *"
        list="origins-list"
        value={formData.origin || ""}
        onChange={(e) => setFormData({ ...formData, origin: e.target.value })}
      />
      <datalist id="origins-list">
        {uniqueOrigins.map((origin) => (
          <option key={origin} value={origin} />
        ))}
      </datalist>
      <input
        type="text"
        className="pokemon-input mb-sm"
        placeholder="Roaster"
        list="roasters-list"
        value={formData.roaster || ""}
        onChange={(e) => setFormData({ ...formData, roaster: e.target.value })}
      />
      <datalist id="roasters-list">
        {uniqueRoasters.map((roaster) => (
          <option key={roaster} value={roaster} />
        ))}
      </datalist>
      <input
        type="text"
        className="pokemon-input"
        placeholder="Variety (e.g., Geisha, Bourbon)"
        value={formData.variety || ""}
        onChange={(e) => setFormData({ ...formData, variety: e.target.value })}
      />
    </div>
  );

  // Step 2: Roast level and processing method
  const renderStep2 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">ROAST & PROCESS (2/2)</div>
      <div className="pokemon-form-label">Roast Level:</div>
      <select
        className="pokemon-select mb-md"
        value={formData.roast_level || "medium"}
        onChange={(e) =>
          setFormData({ ...formData, roast_level: e.target.value as any })
        }
      >
        <option value="light">Light</option>
        <option value="medium">Medium</option>
        <option value="dark">Dark</option>
        <option value="light medium">Light Medium</option>
        <option value="medium dark">Medium Dark</option>
        <option value="unclear">Unclear</option>
      </select>
      <div className="pokemon-form-label">Processing Method:</div>
      <select
        className="pokemon-select"
        value={formData.processing_method || "washed"}
        onChange={(e) =>
          setFormData({
            ...formData,
            processing_method: e.target.value as any,
          })
        }
      >
        <option value="washed">Washed</option>
        <option value="natural">Natural</option>
        <option value="honey">Honey</option>
        <option value="coferment">Coferment</option>
        <option value="experimental">Experimental</option>
      </select>
      <div
        className="pokemon-textbox mt-lg text-center"
        style={{ fontSize: "10px" }}
      >
        After saving, you can log brews to unlock Pokemon!
      </div>
    </div>
  );

  return (
    <div className="pokemon-screen">
      <div
        className="pokemon-frame"
        style={{ maxWidth: "600px", margin: "0 auto" }}
      >
        <button className="pokemon-button mb-md" onClick={onBack}>
          Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
          NEW COFFEE
        </h2>

        {formStep === 1 && renderStep1()}
        {formStep === 2 && renderStep2()}

        {error && (
          <div
            className="pokemon-textbox mt-md"
            style={{ background: "#ffcccc", borderColor: "#cc0000" }}
          >
            <div style={{ fontSize: "10px" }}>{error}</div>
          </div>
        )}

        <div className="pokemon-nav mt-lg">
          {formStep > 1 && (
            <button
              className="pokemon-button"
              onClick={() => setFormStep(formStep - 1)}
            >
              Prev
            </button>
          )}
          <div style={{ flex: 1 }} />
          {formStep < 2 ? (
            <button
              className="pokemon-button"
              onClick={() => setFormStep(formStep + 1)}
            >
              Next
            </button>
          ) : (
            <button className="pokemon-button" onClick={onSubmit}>
              Save Coffee
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default CoffeeForm;
