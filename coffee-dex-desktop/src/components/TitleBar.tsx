import React from "react";

const TitleBar: React.FC = () => {
  return (
    <div
      style={
        {
          position: "fixed",
          top: 0,
          left: 0,
          right: 0,
          height: "32px",
          background: "#181010",
          borderBottom: "2px solid #000",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "0 8px",
          zIndex: 10000,
          WebkitAppRegion: "drag",
          userSelect: "none",
        } as React.CSSProperties
      }
    >
      <div
        style={{
          color: "#9bbc0f",
          fontSize: "10px",
          fontWeight: "bold",
          display: "flex",
          alignItems: "center",
          gap: "8px",
        }}
      >
        <span>☕</span>
        <span>COFFEEDEX</span>
      </div>
      <div
        style={
          {
            display: "flex",
            gap: "4px",
            WebkitAppRegion: "no-drag",
          } as React.CSSProperties
        }
      >
        <button
          onClick={() => {
            (window as any).electron?.minimizeWindow();
          }}
          style={{
            width: "24px",
            height: "24px",
            border: "2px solid #9bbc0f",
            background: "#181010",
            color: "#9bbc0f",
            cursor: "pointer",
            fontSize: "16px",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: 0,
          }}
        >
          −
        </button>
        <button
          onClick={() => {
            (window as any).electron?.closeWindow();
          }}
          style={{
            width: "24px",
            height: "24px",
            border: "2px solid #cc0000",
            background: "#181010",
            color: "#cc0000",
            cursor: "pointer",
            fontSize: "16px",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: 0,
          }}
        >
          ×
        </button>
      </div>
    </div>
  );
};

export default TitleBar;
