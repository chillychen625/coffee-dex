import React from "react";
import "./PokedexFrame.css";

interface PokedexFrameProps {
  children: React.ReactNode;
}

const PokedexFrame: React.FC<PokedexFrameProps> = ({ children }) => {
  return <div className="pokedex-container">{children}</div>;
};

export default PokedexFrame;
