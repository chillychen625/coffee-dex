import React, { useState, useEffect } from "react";
import { Brew, TastingTraits, Coffee } from "../types/pokemon";
import "../styles/pokemon-gameboy.css";
import { api } from "../services/api";

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
  formStep: number;
  setFormStep: (step: number) => void;
  onSubmit: () => void;
  onBack: () => void;
  error: string | null;
}

const BrewForm: React.FC<BrewFormProps> = ({
  coffeeId,
  coffee,
  formData,
  setFormData,
  formStep,
  setFormStep,
  onSubmit,
  onBack,
  error,
}) => {
  const [brewers, setBrewers] = useState<Brewer[]>([]);
  const [uniqueTastingNotes, setUniqueTastingNotes] = useState<string[]>([]);

  useEffect(() => {
    loadBrewers();
    loadTastingNotes();
    // Initialize form data with coffee_id
    if (!formData.coffee_id) {
      setFormData({ ...formData, coffee_id: coffeeId });
    }
  }, [coffeeId]);

  const loadBrewers = async () => {
    try {
      const data = await api.getBrewers();
      setBrewers(data);
    } catch (err) {
      console.error("Failed to load brewers:", err);
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

  // Step 1: Brewer & Rating
  const renderStep1 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">BREW INFO (1/4)</div>
      <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
        <strong>Coffee:</strong> {coffee.name}
        <br />
        <strong>Origin:</strong> {coffee.origin}
      </div>
      <div className="pokemon-form-label">Dripper:</div>
      {brewers.length > 0 ? (
        <select
          className="pokemon-select mb-sm"
          value={formData.dripper || ""}
          onChange={(e) => setFormData({ ...formData, dripper: e.target.value })}
        >
          <option value="">Select brewer...</option>
          {brewers.map((brewer) => (
            <option key={brewer.id} value={brewer.name}>
              {brewer.name}
            </option>
          ))}
        </select>
      ) : (
        <input
          type="text"
          className="pokemon-input mb-sm"
          placeholder="Dripper (e.g., V60)"
          value={formData.dripper || ""}
          onChange={(e) => setFormData({ ...formData, dripper: e.target.value })}
        />
      )}
      <div className="pokemon-form-label">Rating (0-10):</div>
      <input
        type="number"
        className="pokemon-input"
        placeholder="Rating"
        min="0"
        max="10"
        value={formData.rating ?? 5}
        onChange={(e) => setFormData({ ...formData, rating: parseInt(e.target.value) })}
      />
    </div>
  );

  // Step 2: Tasting Notes
  const renderStep2 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">TASTING NOTES (2/4)</div>
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

  // Step 3: Traits
  const renderStep3 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">TASTING TRAITS (3/4)</div>
      {[
        { label: "Sweetness", key: "sweetness" as keyof TastingTraits },
        { label: "Bitterness", key: "bitterness" as keyof TastingTraits },
        { label: "Citrus", key: "citrus_fruits_intensity" as keyof TastingTraits },
        { label: "Berry", key: "berry_intensity" as keyof TastingTraits },
        { label: "Florality", key: "florality" as keyof TastingTraits },
        { label: "Body", key: "body" as keyof TastingTraits },
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
            value={formData.tasting_traits?.[key] ?? 5}
            onChange={(e) => updateTrait(key, parseInt(e.target.value))}
          />
          <div className="pokemon-stat-value">
            {formData.tasting_traits?.[key] ?? 5}
          </div>
        </div>
      ))}
    </div>
  );

  // Step 4: Brew Time
  const renderStep4 = () => (
    <div className="pokemon-form-group">
      <div className="pokemon-subtitle mb-md">BREW TIME (4/4)</div>
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
      <div className="pokemon-textbox mt-md text-center" style={{ fontSize: "10px" }}>
        Ready to save brew!
      </div>
    </div>
  );

  return (
    <div className="pokemon-screen">
      <div className="pokemon-frame" style={{ maxWidth: "600px", margin: "0 auto" }}>
        <button className="pokemon-button mb-md" onClick={onBack}>
          Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
          LOG BREW
        </h2>

        {formStep === 1 && renderStep1()}
        {formStep === 2 && renderStep2()}
        {formStep === 3 && renderStep3()}
        {formStep === 4 && renderStep4()}

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
          {formStep < 4 ? (
            <button
              className="pokemon-button"
              onClick={() => setFormStep(formStep + 1)}
            >
              Next
            </button>
          ) : (
            <button className="pokemon-button" onClick={onSubmit}>
              Save Brew
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default BrewForm;
