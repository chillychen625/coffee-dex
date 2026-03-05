import React from "react";

type EmptyVariant = "no-coffees" | "no-pokemon" | "no-brews" | "no-stats" | "error";

interface EmptyStateProps {
  variant: EmptyVariant;
  title?: string;
  message?: string;
  children?: React.ReactNode;
}

// Map variants to Pokemon and messages
const EMPTY_CONFIG: Record<EmptyVariant, { sprite: string; pokemon: string; title: string; message: string }> = {
  "no-coffees": {
    sprite: "./pokemon-sprites/animated/54.gif", // Psyduck
    pokemon: "Psyduck",
    title: "No Coffees Yet!",
    message: "Even Psyduck needs caffeine. Add your first coffee to get started!",
  },
  "no-pokemon": {
    sprite: "./pokemon-sprites/animated/133.gif", // Eevee
    pokemon: "Eevee",
    title: "No Pokemon Yet!",
    message: "Log 5 brews of a coffee to discover which Pokemon it evolves into!",
  },
  "no-brews": {
    sprite: "./pokemon-sprites/animated/7.gif", // Squirtle
    pokemon: "Squirtle",
    title: "No Brews Yet!",
    message: "Squirtle is ready to help you brew! Log your first brew to start tracking.",
  },
  "no-stats": {
    sprite: "./pokemon-sprites/animated/79.gif", // Slowpoke
    pokemon: "Slowpoke",
    title: "No Statistics Yet!",
    message: "Slowpoke is still gathering data. Start brewing to see your stats!",
  },
  "error": {
    sprite: "./pokemon-sprites/animated/104.gif", // Cubone
    pokemon: "Cubone",
    title: "Something Went Wrong",
    message: "Cubone is sad. Please try again!",
  },
};

const EmptyState: React.FC<EmptyStateProps> = ({
  variant,
  title,
  message,
  children,
}) => {
  const config = EMPTY_CONFIG[variant];

  return (
    <div
      className="pokemon-textbox"
      style={{
        textAlign: "center",
        padding: "24px 16px",
      }}
    >
      <div
        style={{
          width: "80px",
          height: "80px",
          margin: "0 auto 12px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <img
          src={config.sprite}
          alt={config.pokemon}
          style={{
            width: "80px",
            height: "auto",
            imageRendering: "pixelated",
          }}
        />
      </div>
      <div
        style={{
          fontSize: "11px",
          fontWeight: "bold",
          marginBottom: "8px",
        }}
      >
        {title || config.title}
      </div>
      <div
        style={{
          fontSize: "9px",
          lineHeight: "1.4",
          opacity: 0.9,
        }}
      >
        {message || config.message}
      </div>
      {children && <div style={{ marginTop: "12px" }}>{children}</div>}
    </div>
  );
};

export default EmptyState;
