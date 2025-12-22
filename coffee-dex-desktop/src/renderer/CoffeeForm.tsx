import React, { useState, useEffect } from "react";
import { Coffee, TastingTraits } from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import { api } from "../services/api";

interface Brewer {
  id: string;
  name: string;
  pokeball_type: string;
  created_at: string;
}

interface CoffeeFormProps {
  formData: Partial<Coffee>;
  setFormData: (data: Partial<Coffee>) => void;
  formStep: number;
  setFormStep: (step: number) => void;
  onSubmit: () => void;
  onBack: () => void;
  error: string | null;
  isQuickBrew?: boolean;
  recentCoffees?: Coffee[];
  selectedBrewerId?: string;
  onBrewerSelect?: (brewerId: string) => void;
}

const CoffeeForm: React.FC<CoffeeFormProps> = ({
  formData,
  setFormData,
  formStep,
  setFormStep,
  onSubmit,
  onBack,
  error,
  isQuickBrew = false,
  recentCoffees = [],
  selectedBrewerId = "",
  onBrewerSelect = () => {},
}) => {
  const [brewers, setBrewers] = useState<Brewer[]>([]);
  const [allCoffees, setAllCoffees] = useState<Coffee[]>([]);
  const [uniqueOrigins, setUniqueOrigins] = useState<string[]>([]);
  const [uniqueRoasters, setUniqueRoasters] = useState<string[]>([]);
  const [uniqueTastingNotes, setUniqueTastingNotes] = useState<string[]>([]);

  useEffect(() => {
    loadBrewers();
    loadAllCoffeesForAutocomplete();
  }, []);

  const loadBrewers = async () => {
    try {
      const data = await api.getBrewers();
      setBrewers(data);
    } catch (err) {
      console.error("Failed to load brewers:", err);
    }
  };

  const loadAllCoffeesForAutocomplete = async () => {
    try {
      const coffees = await api.getCoffees();
      setAllCoffees(coffees);

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

      // Extract unique tasting notes
      const notes = new Set<string>();
      coffees.forEach((c) => {
        if (c.tasting_notes) {
          c.tasting_notes.forEach((note) => {
            if (note && note.trim()) notes.add(note.trim());
          });
        }
      });
      setUniqueTastingNotes(Array.from(notes).sort());
    } catch (err) {
      console.error("Failed to load coffees for autocomplete:", err);
    }
  };
  const updateTrait = (trait: keyof TastingTraits, value: number) => {
    setFormData({
      ...formData,
      tasting_traits: {
        ...formData.tasting_traits!,
        [trait]: value,
      },
    });
  };

  const updateTastingNote = (index: number, value: string) => {
    const notes = [...(formData.tasting_notes || ["", "", "", "", ""])] as [
      string,
      string,
      string,
      string,
      string
    ];
    notes[index] = value;
    setFormData({ ...formData, tasting_notes: notes });
  };

  const renderStep1 = () => {
    if (isQuickBrew) {
      // Quick Brew: Select from recent coffees
      return (
        <div className="pokemon-form-group">
          <div className="pokemon-subtitle mb-md">QUICK BREW (1/3)</div>
          <div className="pokemon-form-label">Select Coffee:</div>
          <select
            className="pokemon-select mb-sm"
            value={formData.id || ""}
            onChange={(e) => {
              const selected = recentCoffees.find(
                (c) => c.id === e.target.value
              );
              if (selected) {
                setFormData({
                  ...selected,
                  // Reset brew-specific fields
                  tasting_notes: ["", "", "", "", ""],
                  dripper: "",
                  end_time: { minutes: 0, seconds: 0 },
                });
              }
            }}
          >
            <option value="">-- Select a coffee --</option>
            {recentCoffees.map((coffee) => (
              <option key={coffee.id} value={coffee.id}>
                {coffee.name} - {coffee.origin}
              </option>
            ))}
          </select>
          {formData.id && (
            <div className="pokemon-textbox mt-sm" style={{ fontSize: "9px" }}>
              <div>
                <strong>Roaster:</strong> {formData.roaster}
              </div>
              <div>
                <strong>Roast:</strong> {formData.roast_level}
              </div>
              <div>
                <strong>Process:</strong> {formData.processing_method}
              </div>
            </div>
          )}
        </div>
      );
    }

    // Regular new coffee flow
    return (
      <div className="pokemon-form-group">
        <div className="pokemon-subtitle mb-md">BASIC INFO (1/6)</div>
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
          onChange={(e) =>
            setFormData({ ...formData, roaster: e.target.value })
          }
        />
        <datalist id="roasters-list">
          {uniqueRoasters.map((roaster) => (
            <option key={roaster} value={roaster} />
          ))}
        </datalist>
        <input
          type="text"
          className="pokemon-input mb-sm"
          placeholder="Variety (e.g., Geisha, Bourbon)"
          value={formData.variety || ""}
          onChange={(e) =>
            setFormData({ ...formData, variety: e.target.value })
          }
        />
        <input
          type="number"
          className="pokemon-input"
          placeholder="Rating (0-10)"
          min="0"
          max="10"
          value={formData.rating || 5}
          onChange={(e) =>
            setFormData({ ...formData, rating: parseInt(e.target.value) })
          }
        />
      </div>
    );
  };

  const renderStep2 = () => {
    if (isQuickBrew) {
      // Quick Brew: Skip to brewer selection
      return (
        <div className="pokemon-form-group">
          <div className="pokemon-subtitle mb-md">BREWING METHOD (2/3)</div>
          {brewers.length > 0 ? (
            <select
              className="pokemon-select"
              value={selectedBrewerId}
              onChange={(e) => onBrewerSelect(e.target.value)}
              style={{
                width: "100%",
                padding: "8px",
                fontSize: "12px",
                border: "2px solid #000",
              }}
            >
              <option value="">Brewer (optional)</option>
              {brewers.map((brewer) => (
                <option key={brewer.id} value={brewer.id}>
                  {brewer.name}
                </option>
              ))}
            </select>
          ) : (
            <input
              type="text"
              className="pokemon-input"
              placeholder="Dripper (e.g., V60)"
              value={formData.dripper || ""}
              onChange={(e) =>
                setFormData({ ...formData, dripper: e.target.value })
              }
            />
          )}
        </div>
      );
    }

    return (
      <div className="pokemon-form-group">
        <div className="pokemon-subtitle mb-md">ROAST & PROCESS (2/6)</div>
        <select
          className="pokemon-select mb-sm"
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
        <select
          className="pokemon-select mb-sm"
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
        {brewers.length > 0 ? (
          <select
            className="pokemon-select"
            value={selectedBrewerId}
            onChange={(e) => onBrewerSelect(e.target.value)}
            style={{
              width: "100%",
              padding: "8px",
              fontSize: "12px",
              border: "2px solid #000",
            }}
          >
            <option value="">Brewer (optional)</option>
            {brewers.map((brewer) => (
              <option key={brewer.id} value={brewer.id}>
                {brewer.name}
              </option>
            ))}
          </select>
        ) : (
          <input
            type="text"
            className="pokemon-input"
            placeholder="Dripper (e.g., V60)"
            value={formData.dripper || ""}
            onChange={(e) =>
              setFormData({ ...formData, dripper: e.target.value })
            }
          />
        )}
      </div>
    );
  };

  const renderStep3 = () => {
    const stepLabel = isQuickBrew ? "BREW TIME (3/3)" : "TASTING NOTES (3/6)";

    if (isQuickBrew) {
      // Quick Brew: Only tasting notes and brew time
      return (
        <div className="pokemon-form-group">
          <div className="pokemon-subtitle mb-md">{stepLabel}</div>
          <div className="pokemon-form-label">Tasting notes (optional):</div>
          {[0, 1, 2, 3, 4].map((i) => (
            <input
              key={i}
              type="text"
              className="pokemon-input mb-sm"
              placeholder={`Note ${i + 1}`}
              value={formData.tasting_notes?.[i] || ""}
              onChange={(e) => updateTastingNote(i, e.target.value)}
            />
          ))}
          <div className="pokemon-form-row mt-md">
            <div className="pokemon-form-col">
              <label className="pokemon-form-label">Minutes</label>
              <input
                type="number"
                className="pokemon-input"
                placeholder="0"
                min="0"
                value={formData.end_time?.minutes || 0}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    end_time: {
                      ...formData.end_time!,
                      minutes: parseInt(e.target.value) || 0,
                    },
                  })
                }
              />
            </div>
            <div className="pokemon-form-col">
              <label className="pokemon-form-label">Seconds</label>
              <input
                type="number"
                className="pokemon-input"
                placeholder="0"
                min="0"
                max="59"
                value={formData.end_time?.seconds || 0}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    end_time: {
                      ...formData.end_time!,
                      seconds: parseInt(e.target.value) || 0,
                    },
                  })
                }
              />
            </div>
          </div>
          <div
            className="pokemon-textbox mt-md text-center"
            style={{ fontSize: "10px" }}
          >
            Ready to save brew entry!
          </div>
        </div>
      );
    }

    return (
      <div className="pokemon-form-group">
        <div className="pokemon-subtitle mb-md">TASTING NOTES (3/6)</div>
        <div className="pokemon-form-label">Up to 5 notes:</div>
        {[0, 1, 2, 3, 4].map((i) => (
          <input
            key={i}
            type="text"
            className="pokemon-input mb-sm"
            placeholder={`Note ${i + 1}`}
            list="tasting-notes-list"
            value={formData.tasting_notes?.[i] || ""}
            onChange={(e) => updateTastingNote(i, e.target.value)}
          />
        ))}
        <datalist id="tasting-notes-list">
          {uniqueTastingNotes.map((note) => (
            <option key={note} value={note} />
          ))}
        </datalist>
      </div>
    );
  };

  const renderStep4 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">TRAITS 1/2 (4/6)</div>
      {[
        { label: "Sweetness", key: "sweetness" as keyof TastingTraits },
        { label: "Bitterness", key: "bitterness" as keyof TastingTraits },
        {
          label: "Citrus",
          key: "citrus_fruits_intensity" as keyof TastingTraits,
        },
        { label: "Berry", key: "berry_intensity" as keyof TastingTraits },
        {
          label: "Stonefruit",
          key: "stonefruit_intensity" as keyof TastingTraits,
        },
        { label: "Florality", key: "florality" as keyof TastingTraits },
      ].map(({ label, key }) => (
        <div key={key} className="pokemon-stat-row mb-sm">
          <div className="pokemon-stat-label" style={{ minWidth: "80px" }}>
            {label}:
          </div>
          <input
            type="range"
            className="pokemon-slider"
            min="0"
            max="10"
            value={formData.tasting_traits?.[key] || 5}
            onChange={(e) => updateTrait(key, parseInt(e.target.value))}
          />
          <div className="pokemon-stat-value">
            {formData.tasting_traits?.[key] || 5}
          </div>
        </div>
      ))}
    </div>
  );

  const renderStep5 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">TRAITS 2/2 (5/6)</div>
      {[
        { label: "Roast", key: "roast_intensity" as keyof TastingTraits },
        { label: "Spice", key: "spice" as keyof TastingTraits },
        { label: "Aromatic", key: "aromatic_intensity" as keyof TastingTraits },
        { label: "Savory", key: "savory" as keyof TastingTraits },
        { label: "Body", key: "body" as keyof TastingTraits },
        { label: "Cleanliness", key: "cleanliness" as keyof TastingTraits },
      ].map(({ label, key }) => (
        <div key={key} className="pokemon-stat-row mb-sm">
          <div className="pokemon-stat-label" style={{ minWidth: "80px" }}>
            {label}:
          </div>
          <input
            type="range"
            className="pokemon-slider"
            min="0"
            max="10"
            value={formData.tasting_traits?.[key] || 5}
            onChange={(e) => updateTrait(key, parseInt(e.target.value))}
          />
          <div className="pokemon-stat-value">
            {formData.tasting_traits?.[key] || 5}
          </div>
        </div>
      ))}
    </div>
  );

  const renderStep6 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">BREW TIME (6/6)</div>
      <div className="pokemon-form-row">
        <div className="pokemon-form-col">
          <label className="pokemon-form-label">Minutes</label>
          <input
            type="number"
            className="pokemon-input"
            placeholder="0"
            min="0"
            value={formData.end_time?.minutes || 0}
            onChange={(e) =>
              setFormData({
                ...formData,
                end_time: {
                  ...formData.end_time!,
                  minutes: parseInt(e.target.value) || 0,
                },
              })
            }
          />
        </div>
        <div className="pokemon-form-col">
          <label className="pokemon-form-label">Seconds</label>
          <input
            type="number"
            className="pokemon-input"
            placeholder="0"
            min="0"
            max="59"
            value={formData.end_time?.seconds || 0}
            onChange={(e) =>
              setFormData({
                ...formData,
                end_time: {
                  ...formData.end_time!,
                  seconds: parseInt(e.target.value) || 0,
                },
              })
            }
          />
        </div>
      </div>
      <div
        className="pokemon-textbox mt-md text-center"
        style={{ fontSize: "10px" }}
      >
        Ready to generate Pokemon!
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
          ← Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
          {isQuickBrew ? "QUICK BREW" : "NEW COFFEE"}
        </h2>

        {formStep === 1 && renderStep1()}
        {formStep === 2 && renderStep2()}
        {formStep === 3 && renderStep3()}
        {!isQuickBrew && formStep === 4 && renderStep4()}
        {!isQuickBrew && formStep === 5 && renderStep5()}
        {!isQuickBrew && formStep === 6 && renderStep6()}

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
              ← Prev
            </button>
          )}
          <div style={{ flex: 1 }} />
          {formStep < (isQuickBrew ? 3 : 6) ? (
            <button
              className="pokemon-button"
              onClick={() => setFormStep(formStep + 1)}
              disabled={isQuickBrew && formStep === 1 && !formData.id}
            >
              Next →
            </button>
          ) : (
            <button className="pokemon-button" onClick={onSubmit}>
              {isQuickBrew ? "Save Brew" : "Generate!"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default CoffeeForm;
