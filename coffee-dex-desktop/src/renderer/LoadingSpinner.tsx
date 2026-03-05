import React from "react";

type LoadingVariant = "default" | "pokedex" | "brew" | "save" | "generate";

interface LoadingSpinnerProps {
  message?: string;
  variant?: LoadingVariant;
}

// Map variants to Pokemon sprites and messages
const LOADING_CONFIG: Record<LoadingVariant, { sprite: string; pokemon: string; defaultMessage: string }> = {
  default: {
    sprite: "./pokemon-sprites/animated/79.gif", // Slowpoke
    pokemon: "Slowpoke",
    defaultMessage: "Loading...",
  },
  pokedex: {
    sprite: "./pokemon-sprites/animated/143.gif", // Snorlax
    pokemon: "Snorlax",
    defaultMessage: "Loading Pokedex...",
  },
  brew: {
    sprite: "./pokemon-sprites/animated/7.gif", // Squirtle
    pokemon: "Squirtle",
    defaultMessage: "Brewing...",
  },
  save: {
    sprite: "./pokemon-sprites/animated/25.gif", // Pikachu
    pokemon: "Pikachu",
    defaultMessage: "Saving...",
  },
  generate: {
    sprite: "./pokemon-sprites/animated/133.gif", // Eevee
    pokemon: "Eevee",
    defaultMessage: "Generating Pokemon...",
  },
};

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  message,
  variant = "default",
}) => {
  const config = LOADING_CONFIG[variant];

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "24px",
      }}
    >
      <div
        style={{
          position: "relative",
          width: "96px",
          height: "96px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <img
          src={config.sprite}
          alt={config.pokemon}
          style={{
            width: "96px",
            height: "auto",
            imageRendering: "pixelated",
          }}
        />
      </div>
      <div
        className="pokemon-loading"
        style={{
          marginTop: "12px",
          animation: "none", // Override the dots animation from class
        }}
      >
        {message || config.defaultMessage}
      </div>
    </div>
  );
};

export default LoadingSpinner;
