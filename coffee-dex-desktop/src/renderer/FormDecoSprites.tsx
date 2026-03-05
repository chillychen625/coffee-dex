import React, { useMemo } from "react";

interface FormDecoSpritesProps {
  seed?: string; // Optional seed for consistent randomness per form instance
  spin?: boolean; // Make the sprites spin (for start screen fun)
}

// Generate a random Pokemon number from 1-151
const getRandomPokemon = (seed: number): number => {
  // Simple seeded random for consistency
  const x = Math.sin(seed) * 10000;
  return Math.floor((x - Math.floor(x)) * 151) + 1;
};

// Pad number to 3 digits (e.g., 1 -> "001", 25 -> "025")
const padPokemonId = (id: number): string => {
  return id.toString().padStart(3, "0");
};

const FormDecoSprites: React.FC<FormDecoSpritesProps> = ({ seed = "default", spin = false }) => {
  // Generate 4 different Pokemon for corners based on seed
  const cornerPokemon = useMemo(() => {
    const baseSeed = seed.split("").reduce((a, c) => a + c.charCodeAt(0), 0);
    return [
      getRandomPokemon(baseSeed),
      getRandomPokemon(baseSeed + 1),
      getRandomPokemon(baseSeed + 2),
      getRandomPokemon(baseSeed + 3),
    ];
  }, [seed]);

  const cornerPositions = [
    { top: "8px", left: "8px" },      // Top-left
    { top: "8px", right: "8px" },     // Top-right
    { bottom: "8px", left: "8px" },   // Bottom-left
    { bottom: "8px", right: "8px" },  // Bottom-right
  ];

  // Different spin speeds for each corner (more silly!)
  const spinDurations = ["3s", "4s", "5s", "3.5s"];

  return (
    <>
      {spin && (
        <style>{`
          @keyframes pokemon-spin {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
          }
        `}</style>
      )}
      {cornerPokemon.map((pokemonId, index) => (
        <img
          key={index}
          src={`./pokemon-sprites/${padPokemonId(pokemonId)}.png`}
          alt=""
          aria-hidden="true"
          style={{
            position: "absolute",
            width: "80px",
            height: "80px",
            opacity: 0.7,
            imageRendering: "pixelated",
            pointerEvents: "none",
            zIndex: 0,
            ...(spin && {
              animation: `pokemon-spin ${spinDurations[index]} linear infinite`,
            }),
            ...cornerPositions[index],
          }}
        />
      ))}
    </>
  );
};

export default FormDecoSprites;
