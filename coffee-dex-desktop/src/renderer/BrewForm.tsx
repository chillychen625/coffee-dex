import React, { useState, useEffect } from "react";
import { Brew, TastingTraits, Coffee } from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import { api } from "../services/api";
import FormDecoSprites from "./FormDecoSprites";

interface Brewer {
  id: string;
  name: string;
  pokeball_type: string;
  created_at: string;
}

interface BrewFormProps {
  coffeeId: string;
  coffee: Coffee;
  formData: Partial<Brew>;
  setFormData: (data: Partial<Brew>) => void;
  onSubmit: () => void;
  onBack: () => void;
  error: string | null;
}

const BrewForm: React.FC<BrewFormProps> = ({
  coffeeId,
  coffee,
  formData,
  setFormData,
  onSubmit,
  onBack,
  error,
}) => {
  const [brewers, setBrewers] = useState<Brewer[]>([]);
  const [uniqueTastingNotes, setUniqueTastingNotes] = useState<string[]>([]);
  const [traitsExpanded, setTraitsExpanded] = useState(false);

  useEffect(() => {
    loadBrewers();
    loadTastingNotes();
    if (!formData.coffee_id) {
      setFormData({ ...formData, coffee_id: coffeeId });
    }
  }, [coffeeId]);

  const loadBrewers = async () => {
    try {
      const data = await api.getBrewers();
      setBrewers(data || []);
    } catch (err) {
      console.error("Failed to load brewers:", err);
      setBrewers([]);
    }
  };

  const loadTastingNotes = async () => {
    try {
      const brews = await api.getBrews();
      const notes = new Set<string>();
      brews.forEach((b) => {
        if (b.tasting_notes) {
          b.tasting_notes.forEach((note) => {
            if (note && note.trim()) notes.add(note.trim());
          });
        }
      });
      setUniqueTastingNotes(Array.from(notes).sort());
    } catch (err) {
      console.error("Failed to load tasting notes:", err);
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

  const ALL_TRAITS: { label: string; key: keyof TastingTraits }[] = [
    { label: "Sweetness", key: "sweetness" },
    { label: "Bitterness", key: "bitterness" },
    { label: "Body", key: "body" },
    { label: "Citrus", key: "citrus_fruits_intensity" },
    { label: "Berry", key: "berry_intensity" },
    { label: "Florality", key: "florality" },
    { label: "Stonefruit", key: "stonefruit_intensity" },
    { label: "Roast", key: "roast_intensity" },
    { label: "Aromatic", key: "aromatic_intensity" },
    { label: "Spice", key: "spice" },
    { label: "Savory", key: "savory" },
    { label: "Cleanliness", key: "cleanliness" },
  ];

  const toggleTrait = (key: keyof TastingTraits) => {
    const current = formData.tasting_traits?.[key] ?? -1;
    if (current === -1) {
      updateTrait(key, 5);
    } else {
      updateTrait(key, -1);
    }
  };

  const brewerList = brewers || [];
  const scoredCount = ALL_TRAITS.filter(
    ({ key }) => (formData.tasting_traits?.[key] ?? -1) >= 0
  ).length;

  return (
    <div className="pokemon-screen">
      <div className="pokemon-frame" style={{ position: "relative" }}>
        <FormDecoSprites seed={`brew-${coffeeId}`} spin={true} />
        <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 1 }}>
          Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
          LOG BREW
        </h2>

        {/* ── BREW SETUP ── */}
        <div className="pokemon-form-group">
          <div className="pokemon-subtitle mb-md">BREW SETUP</div>
          <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
            <strong>Coffee:</strong> {coffee.name}
            <br />
            <strong>Origin:</strong> {coffee.origin}
          </div>
          <div className="pokemon-form-label">Dripper:</div>
          {brewerList.length > 0 ? (
            <select
              className="pokemon-select mb-sm"
              value={formData.dripper || ""}
              onChange={(e) => setFormData({ ...formData, dripper: e.target.value })}
            >
              <option value="">Select brewer...</option>
              {brewerList.map((brewer) => (
                <option key={brewer.id} value={brewer.name}>
                  {brewer.name}
                </option>
              ))}
            </select>
          ) : (
            <>
              <input
                type="text"
                className="pokemon-input mb-sm"
                placeholder="Dripper (e.g., V60)"
                value={formData.dripper || ""}
                onChange={(e) => setFormData({ ...formData, dripper: e.target.value })}
              />
              <div
                className="pokemon-textbox mb-sm"
                style={{ fontSize: "9px", background: "#ffffcc", borderColor: "#aaaa00" }}
              >
                Tip: Set up brewers in the Brewers menu for quick selection!
              </div>
            </>
          )}
          <div className="pokemon-form-label">Rating (0-10):</div>
          <input
            type="number"
            className="pokemon-input mb-md"
            placeholder="Rating"
            min="0"
            max="10"
            value={formData.rating ?? 5}
            onChange={(e) => setFormData({ ...formData, rating: parseInt(e.target.value) })}
          />
          <div className="pokemon-form-label">Brew Time:</div>
          <div className="pokemon-form-row">
            <div className="pokemon-form-col">
              <label className="pokemon-form-label" style={{ fontSize: "9px" }}>Minutes</label>
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
              <label className="pokemon-form-label" style={{ fontSize: "9px" }}>Seconds</label>
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
        </div>

        {/* ── TASTING NOTES ── */}
        <div className="pokemon-form-group mt-lg">
          <div className="pokemon-subtitle mb-md">TASTING NOTES</div>
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

        {/* ── TRAIT SCORING (collapsible) ── */}
        <div className="pokemon-form-group mt-lg">
          <div
            className="pokemon-subtitle mb-md"
            style={{ cursor: "pointer", display: "flex", alignItems: "center", gap: "6px" }}
            onClick={() => setTraitsExpanded(!traitsExpanded)}
          >
            TRAIT SCORING
            <span style={{ fontSize: "9px", opacity: 0.6 }}>
              {traitsExpanded ? "[ - ]" : "[ + ]"} {scoredCount > 0 ? `(${scoredCount}/12 scored)` : "(optional)"}
            </span>
          </div>
          {traitsExpanded && (
            <>
              <div
                className="pokemon-textbox mb-sm"
                style={{ fontSize: "8px", padding: "4px 8px", background: "#ffffcc", borderColor: "#aaaa00" }}
              >
                Toggle on traits you want to score. Leave off to skip.
              </div>
              {ALL_TRAITS.map(({ label, key }) => {
                const value = formData.tasting_traits?.[key] ?? -1;
                const isEnabled = value >= 0;
                return (
                  <div key={key} className="pokemon-stat-row mb-sm" style={{ alignItems: "center" }}>
                    <label
                      style={{
                        display: "flex",
                        alignItems: "center",
                        minWidth: "100px",
                        cursor: "pointer",
                        fontSize: "10px",
                        gap: "4px",
                        opacity: isEnabled ? 1 : 0.5,
                      }}
                    >
                      <input
                        type="checkbox"
                        checked={isEnabled}
                        onChange={() => toggleTrait(key)}
                        style={{ cursor: "pointer" }}
                      />
                      {label}
                    </label>
                    {isEnabled ? (
                      <>
                        <input
                          type="range"
                          className="pokemon-slider"
                          min="0"
                          max="10"
                          value={value}
                          onChange={(e) => updateTrait(key, parseInt(e.target.value))}
                          style={{ flex: 1 }}
                        />
                        <div className="pokemon-stat-value" style={{ minWidth: "20px", textAlign: "right" }}>
                          {value}
                        </div>
                      </>
                    ) : (
                      <div style={{ flex: 1, fontSize: "9px", opacity: 0.4, paddingLeft: "8px" }}>
                        --
                      </div>
                    )}
                  </div>
                );
              })}
            </>
          )}
        </div>

        {error && (
          <div
            className="pokemon-textbox mt-md"
            style={{ background: "#ffcccc", borderColor: "#cc0000" }}
          >
            <div style={{ fontSize: "10px" }}>{error}</div>
          </div>
        )}

        <div className="pokemon-nav mt-lg">
          <div style={{ flex: 1 }} />
          <button className="pokemon-button" onClick={onSubmit}>
            Save Brew
          </button>
        </div>
      </div>
    </div>
  );
};

export default BrewForm;
