import React, { useState, useEffect } from "react";
import { api } from "../services/api";
import { Coffee, Brew, TastingTraits } from "../types/pokemon";
import LoadingSpinner from "./LoadingSpinner";
import FormDecoSprites from "./FormDecoSprites";

interface CoffeeComparisonProps {
  onBack: () => void;
}

// Helper to average a single trait across brews, skipping -1 (not scored)
const avgTrait = (brews: Brew[], key: keyof TastingTraits): number => {
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

// Helper to compute average traits from brews (skips -1 / not scored)
const computeAverageTraits = (brews: Brew[]): TastingTraits | null => {
  if (brews.length === 0) return null;

  const traitKeys: (keyof TastingTraits)[] = [
    "berry_intensity", "stonefruit_intensity", "roast_intensity",
    "citrus_fruits_intensity", "bitterness", "florality", "spice",
    "sweetness", "aromatic_intensity", "savory", "body", "cleanliness",
  ];

  const result = {} as TastingTraits;
  traitKeys.forEach((key) => {
    result[key] = avgTrait(brews, key);
  });
  return result;
};

// Trait config for the comparison radar chart
const TRAIT_CONFIG = [
  { key: "berry_intensity", label: "Berry", color: "#c03028" },
  { key: "stonefruit_intensity", label: "Stone", color: "#f08030" },
  { key: "citrus_fruits_intensity", label: "Citrus", color: "#f8d030" },
  { key: "florality", label: "Floral", color: "#78c850" },
  { key: "aromatic_intensity", label: "Aroma", color: "#98d8d8" },
  { key: "cleanliness", label: "Clean", color: "#6890f0" },
  { key: "sweetness", label: "Sweet", color: "#f85888" },
  { key: "body", label: "Body", color: "#a040a0" },
  { key: "bitterness", label: "Bitter", color: "#705848" },
  { key: "roast_intensity", label: "Roast", color: "#b8a038" },
  { key: "savory", label: "Savory", color: "#e0c068" },
  { key: "spice", label: "Spice", color: "#a890f0" },
] as const;

// Comparison radar chart - overlays two profiles
const ComparisonWheel: React.FC<{
  traits1: TastingTraits | null;
  traits2: TastingTraits | null;
  size?: number;
}> = ({ traits1, traits2, size = 220 }) => {
  const center = size / 2;
  const maxRadius = size / 2 - 35;
  const numAxes = TRAIT_CONFIG.length;
  const angleStep = (2 * Math.PI) / numAxes;

  const getPoint = (index: number, value: number): { x: number; y: number } => {
    const angle = index * angleStep - Math.PI / 2;
    const radius = (value / 10) * maxRadius;
    return {
      x: center + radius * Math.cos(angle),
      y: center + radius * Math.sin(angle),
    };
  };

  const generatePath = (traits: TastingTraits | null): string => {
    if (!traits) return "";
    const points = TRAIT_CONFIG.map((trait, i) => {
      const value = Math.max(0, traits[trait.key as keyof TastingTraits] ?? 0);
      const point = getPoint(i, value);
      return `${point.x},${point.y}`;
    });
    return `M ${points.join(" L ")} Z`;
  };

  const generateAxisLines = () => {
    return TRAIT_CONFIG.map((_, i) => {
      const endPoint = getPoint(i, 10);
      return (
        <line
          key={`axis-${i}`}
          x1={center}
          y1={center}
          x2={endPoint.x}
          y2={endPoint.y}
          stroke="var(--gb-dark)"
          strokeWidth="1"
          opacity="0.3"
        />
      );
    });
  };

  const generateScaleCircles = () => {
    const circles = [];
    for (let i = 2; i <= 10; i += 2) {
      const radius = (i / 10) * maxRadius;
      circles.push(
        <circle
          key={`scale-${i}`}
          cx={center}
          cy={center}
          r={radius}
          fill="none"
          stroke="var(--gb-dark)"
          strokeWidth="1"
          opacity="0.2"
        />
      );
    }
    return circles;
  };

  const generateLabels = () => {
    return TRAIT_CONFIG.map((trait, i) => {
      const angle = i * angleStep - Math.PI / 2;
      const labelRadius = maxRadius + 20;
      const x = center + labelRadius * Math.cos(angle);
      const y = center + labelRadius * Math.sin(angle);

      return (
        <text
          key={`label-${i}`}
          x={x}
          y={y}
          textAnchor="middle"
          dominantBaseline="middle"
          fontSize="7"
          fontFamily="'Press Start 2P', monospace"
          fill="var(--gb-darkest)"
        >
          {trait.label}
        </text>
      );
    });
  };

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
      <circle
        cx={center}
        cy={center}
        r={maxRadius}
        fill="var(--gb-lightest)"
        opacity="0.3"
      />
      {generateScaleCircles()}
      {generateAxisLines()}

      {/* Coffee 1 - Blue */}
      {traits1 && (
        <path
          d={generatePath(traits1)}
          fill="rgba(104, 144, 240, 0.3)"
          stroke="#6890f0"
          strokeWidth="2"
        />
      )}

      {/* Coffee 2 - Red */}
      {traits2 && (
        <path
          d={generatePath(traits2)}
          fill="rgba(240, 128, 48, 0.3)"
          stroke="#f08030"
          strokeWidth="2"
        />
      )}

      {generateLabels()}
    </svg>
  );
};

interface CoffeeWithBrews {
  coffee: Coffee;
  brews: Brew[];
  avgRating: number;
  avgTraits: TastingTraits | null;
}

const CoffeeComparison: React.FC<CoffeeComparisonProps> = ({ onBack }) => {
  const [coffees, setCoffees] = useState<Coffee[]>([]);
  const [loading, setLoading] = useState(true);
  const [coffee1, setCoffee1] = useState<CoffeeWithBrews | null>(null);
  const [coffee2, setCoffee2] = useState<CoffeeWithBrews | null>(null);
  const [selecting, setSelecting] = useState<1 | 2 | null>(null);

  useEffect(() => {
    loadCoffees();
  }, []);

  const loadCoffees = async () => {
    try {
      const data = await api.getCoffees();
      setCoffees(data);
    } catch (error) {
      console.error("Failed to load coffees:", error);
    } finally {
      setLoading(false);
    }
  };

  const selectCoffee = async (coffee: Coffee, slot: 1 | 2) => {
    try {
      const brews = await api.getBrewsForCoffee(coffee.id);
      const avgTraits = computeAverageTraits(brews);
      const avgRating =
        brews.length > 0
          ? brews.reduce((sum, b) => sum + b.rating, 0) / brews.length
          : 0;

      const coffeeWithBrews: CoffeeWithBrews = {
        coffee,
        brews,
        avgRating,
        avgTraits,
      };

      if (slot === 1) {
        setCoffee1(coffeeWithBrews);
      } else {
        setCoffee2(coffeeWithBrews);
      }
      setSelecting(null);
    } catch (error) {
      console.error("Failed to load coffee details:", error);
    }
  };

  if (loading) {
    return (
      <div className="pokemon-screen centered">
        <div className="pokemon-frame" style={{ position: "relative" }}>
          <FormDecoSprites seed="comparison-loading" spin={true} />
          <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 1 }}>
            Back
          </button>
          <LoadingSpinner variant="default" message="Loading coffees..." />
        </div>
      </div>
    );
  }

  // Selection screen
  if (selecting !== null) {
    return (
      <div className="pokemon-screen">
        <div className="pokemon-frame" style={{ position: "relative" }}>
          <FormDecoSprites seed={`select-${selecting}`} spin={true} />
          <button className="pokemon-button mb-md" onClick={() => setSelecting(null)} style={{ position: "relative", zIndex: 1 }}>
            Cancel
          </button>

          <h2 className="pokemon-title" style={{ fontSize: "12px", position: "relative", zIndex: 1 }}>
            SELECT COFFEE {selecting}
          </h2>

          <div>
            {coffees.map((coffee) => (
              <button
                key={coffee.id}
                className="pokemon-textbox mb-sm"
                style={{
                  width: "100%",
                  textAlign: "left",
                  cursor: "pointer",
                  border: "2px solid #000",
                  opacity:
                    (selecting === 1 && coffee2?.coffee.id === coffee.id) ||
                    (selecting === 2 && coffee1?.coffee.id === coffee.id)
                      ? 0.4
                      : 1,
                }}
                onClick={() => {
                  if (
                    (selecting === 1 && coffee2?.coffee.id === coffee.id) ||
                    (selecting === 2 && coffee1?.coffee.id === coffee.id)
                  ) {
                    return; // Can't select the same coffee for both slots
                  }
                  selectCoffee(coffee, selecting);
                }}
              >
                <div style={{ fontWeight: "bold", fontSize: "10px" }}>{coffee.name}</div>
                <div style={{ fontSize: "8px" }}>
                  {coffee.origin} | {coffee.roaster}
                </div>
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  // Helper to compare values
  const compareValues = (v1: number, v2: number): "higher" | "lower" | "equal" => {
    if (Math.abs(v1 - v2) < 0.5) return "equal";
    return v1 > v2 ? "higher" : "lower";
  };

  const getCompareStyle = (comparison: "higher" | "lower" | "equal", forSlot: 1 | 2) => {
    if (comparison === "equal") return {};
    if ((comparison === "higher" && forSlot === 1) || (comparison === "lower" && forSlot === 2)) {
      return { color: "#00aa00", fontWeight: "bold" };
    }
    return { color: "#cc0000" };
  };

  return (
    <div className="pokemon-screen">
      <div className="pokemon-frame" style={{ position: "relative" }}>
        <FormDecoSprites seed="comparison" spin={true} />
        <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 1 }}>
          Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "12px", marginBottom: "12px", position: "relative", zIndex: 1 }}>
          COFFEE COMPARISON
        </h2>

        {/* Selection slots */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px", marginBottom: "12px" }}>
          <button
            className="pokemon-textbox"
            style={{
              cursor: "pointer",
              textAlign: "center",
              border: "2px solid #6890f0",
              backgroundColor: coffee1 ? "rgba(104, 144, 240, 0.1)" : undefined,
            }}
            onClick={() => setSelecting(1)}
          >
            {coffee1 ? (
              <>
                <div style={{ fontSize: "9px", fontWeight: "bold", color: "#6890f0" }}>
                  {coffee1.coffee.name}
                </div>
                <div style={{ fontSize: "7px" }}>{coffee1.coffee.origin}</div>
              </>
            ) : (
              <div style={{ fontSize: "9px", color: "#6890f0" }}>Select Coffee 1</div>
            )}
          </button>

          <button
            className="pokemon-textbox"
            style={{
              cursor: "pointer",
              textAlign: "center",
              border: "2px solid #f08030",
              backgroundColor: coffee2 ? "rgba(240, 128, 48, 0.1)" : undefined,
            }}
            onClick={() => setSelecting(2)}
          >
            {coffee2 ? (
              <>
                <div style={{ fontSize: "9px", fontWeight: "bold", color: "#f08030" }}>
                  {coffee2.coffee.name}
                </div>
                <div style={{ fontSize: "7px" }}>{coffee2.coffee.origin}</div>
              </>
            ) : (
              <div style={{ fontSize: "9px", color: "#f08030" }}>Select Coffee 2</div>
            )}
          </button>
        </div>

        {/* Comparison content */}
        {coffee1 && coffee2 && (
          <>
            {/* Overlaid flavor wheel */}
            <div className="pokemon-textbox mb-sm" style={{ textAlign: "center" }}>
              <div style={{ fontWeight: "bold", fontSize: "9px", marginBottom: "4px" }}>
                FLAVOR PROFILE COMPARISON
              </div>
              <div style={{ display: "flex", justifyContent: "center" }}>
                <ComparisonWheel traits1={coffee1.avgTraits} traits2={coffee2.avgTraits} />
              </div>
              <div style={{ display: "flex", justifyContent: "center", gap: "16px", marginTop: "4px", fontSize: "8px" }}>
                <span style={{ color: "#6890f0" }}>■ {coffee1.coffee.name}</span>
                <span style={{ color: "#f08030" }}>■ {coffee2.coffee.name}</span>
              </div>
            </div>

            {/* Stats comparison */}
            <div className="pokemon-textbox mb-sm" style={{ fontSize: "8px" }}>
              <div style={{ fontWeight: "bold", marginBottom: "8px", textAlign: "center" }}>
                QUICK STATS
              </div>
              <div style={{ display: "grid", gridTemplateColumns: "1fr auto 1fr", gap: "4px" }}>
                {/* Rating */}
                <div style={{ textAlign: "right", ...getCompareStyle(compareValues(coffee1.avgRating, coffee2.avgRating), 1) }}>
                  {coffee1.avgRating.toFixed(1)}/10
                </div>
                <div style={{ textAlign: "center", fontWeight: "bold" }}>Rating</div>
                <div style={{ ...getCompareStyle(compareValues(coffee2.avgRating, coffee1.avgRating), 2) }}>
                  {coffee2.avgRating.toFixed(1)}/10
                </div>

                {/* Brews */}
                <div style={{ textAlign: "right", ...getCompareStyle(compareValues(coffee1.brews.length, coffee2.brews.length), 1) }}>
                  {coffee1.brews.length}
                </div>
                <div style={{ textAlign: "center", fontWeight: "bold" }}>Brews</div>
                <div style={{ ...getCompareStyle(compareValues(coffee2.brews.length, coffee1.brews.length), 2) }}>
                  {coffee2.brews.length}
                </div>

                {/* Origin */}
                <div style={{ textAlign: "right" }}>{coffee1.coffee.origin}</div>
                <div style={{ textAlign: "center", fontWeight: "bold" }}>Origin</div>
                <div>{coffee2.coffee.origin}</div>

                {/* Process */}
                <div style={{ textAlign: "right" }}>{coffee1.coffee.processing_method}</div>
                <div style={{ textAlign: "center", fontWeight: "bold" }}>Process</div>
                <div>{coffee2.coffee.processing_method}</div>

                {/* Roast */}
                <div style={{ textAlign: "right" }}>{coffee1.coffee.roast_level}</div>
                <div style={{ textAlign: "center", fontWeight: "bold" }}>Roast</div>
                <div>{coffee2.coffee.roast_level}</div>
              </div>
            </div>

            {/* Trait-by-trait comparison */}
            {coffee1.avgTraits && coffee2.avgTraits && (
              <div className="pokemon-textbox" style={{ fontSize: "7px" }}>
                <div style={{ fontWeight: "bold", marginBottom: "6px", textAlign: "center" }}>
                  TRAIT COMPARISON
                </div>
                <div style={{ display: "grid", gridTemplateColumns: "1fr auto 1fr", gap: "2px" }}>
                  {TRAIT_CONFIG.map((trait) => {
                    const v1 = Math.max(0, coffee1.avgTraits![trait.key as keyof TastingTraits] ?? 0);
                    const v2 = Math.max(0, coffee2.avgTraits![trait.key as keyof TastingTraits] ?? 0);
                    const cmp = compareValues(v1, v2);
                    return (
                      <React.Fragment key={trait.key}>
                        <div style={{ textAlign: "right", ...getCompareStyle(cmp, 1) }}>{v1}</div>
                        <div style={{ textAlign: "center" }}>{trait.label}</div>
                        <div style={{ ...getCompareStyle(cmp === "higher" ? "lower" : cmp === "lower" ? "higher" : "equal", 2) }}>{v2}</div>
                      </React.Fragment>
                    );
                  })}
                </div>
              </div>
            )}
          </>
        )}

        {/* Prompt to select coffees */}
        {(!coffee1 || !coffee2) && (
          <div className="pokemon-textbox" style={{ textAlign: "center", fontSize: "9px" }}>
            Select two coffees above to compare their flavor profiles!
          </div>
        )}
      </div>
    </div>
  );
};

export default CoffeeComparison;
