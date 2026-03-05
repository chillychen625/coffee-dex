import React from "react";
import { TastingTraits } from "../types/pokemon";

interface FlavorWheelProps {
  traits: TastingTraits;
  size?: number;
  showLabels?: boolean;
}

// Trait labels in display order (clockwise from top)
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

const FlavorWheel: React.FC<FlavorWheelProps> = ({
  traits,
  size = 200,
  showLabels = true,
}) => {
  const center = size / 2;
  const maxRadius = size / 2 - (showLabels ? 30 : 10);
  const numAxes = TRAIT_CONFIG.length;
  const angleStep = (2 * Math.PI) / numAxes;

  // Generate points for each trait value
  const getPoint = (index: number, value: number): { x: number; y: number } => {
    const angle = index * angleStep - Math.PI / 2; // Start from top
    const radius = (value / 10) * maxRadius;
    return {
      x: center + radius * Math.cos(angle),
      y: center + radius * Math.sin(angle),
    };
  };

  // Get trait value, treating -1 (not scored) as 0
  const getTraitValue = (key: keyof TastingTraits): number => {
    const v = traits[key];
    return v != null && v >= 0 ? v : 0;
  };

  // Generate the polygon path for the trait values
  const generatePath = (): string => {
    const points = TRAIT_CONFIG.map((trait, i) => {
      const value = getTraitValue(trait.key as keyof TastingTraits);
      const point = getPoint(i, value);
      return `${point.x},${point.y}`;
    });
    return `M ${points.join(" L ")} Z`;
  };

  // Generate axis lines
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

  // Generate concentric circles for scale
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

  // Generate labels
  const generateLabels = () => {
    if (!showLabels) return null;

    return TRAIT_CONFIG.map((trait, i) => {
      const angle = i * angleStep - Math.PI / 2;
      const labelRadius = maxRadius + 18;
      const x = center + labelRadius * Math.cos(angle);
      const y = center + labelRadius * Math.sin(angle);
      const value = getTraitValue(trait.key as keyof TastingTraits);

      return (
        <text
          key={`label-${i}`}
          x={x}
          y={y}
          textAnchor="middle"
          dominantBaseline="middle"
          fontSize="8"
          fontFamily="'Press Start 2P', monospace"
          fill="var(--gb-darkest)"
        >
          {trait.label}
        </text>
      );
    });
  };

  // Generate value dots at each point
  const generateValueDots = () => {
    return TRAIT_CONFIG.map((trait, i) => {
      const value = getTraitValue(trait.key as keyof TastingTraits);
      const point = getPoint(i, value);
      return (
        <circle
          key={`dot-${i}`}
          cx={point.x}
          cy={point.y}
          r={3}
          fill={trait.color}
          stroke="var(--gb-darkest)"
          strokeWidth="1"
        />
      );
    });
  };

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
      {/* Background */}
      <circle
        cx={center}
        cy={center}
        r={maxRadius}
        fill="var(--gb-lightest)"
        opacity="0.3"
      />

      {/* Scale circles */}
      {generateScaleCircles()}

      {/* Axis lines */}
      {generateAxisLines()}

      {/* Filled area */}
      <path
        d={generatePath()}
        fill="var(--gb-dark)"
        fillOpacity="0.4"
        stroke="var(--gb-darkest)"
        strokeWidth="2"
      />

      {/* Value dots */}
      {generateValueDots()}

      {/* Labels */}
      {generateLabels()}
    </svg>
  );
};

export default FlavorWheel;
