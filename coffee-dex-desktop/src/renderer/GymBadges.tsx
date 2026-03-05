import React, { useState, useEffect } from "react";
import { api } from "../services/api";

interface StatisticsData {
  total_coffees: number;
  total_pokemon: number;
  completion_percent: number;
  average_rating: number;
  highest_rated: {
    name: string;
    rating: number;
  } | null;
  type_distribution: { [key: string]: number };
  origin_distribution: { [key: string]: number };
  processing_stats: {
    [key: string]: {
      count: number;
      average_rating: number;
    };
  };
  brewer_stats: {
    [key: string]: {
      count: number;
    };
  };
}

interface Badge {
  id: number;
  name: string;
  requirement: string;
  emoji: string;
  color: string;
  checkUnlocked: (stats: StatisticsData) => boolean;
  getProgress: (stats: StatisticsData) => { current: number; target: number };
}

const BADGES: Badge[] = [
  {
    id: 1,
    name: "Boulder",
    requirement: "Log 10 coffees",
    emoji: "🪨",
    color: "#a0a0a0",
    checkUnlocked: (stats) => stats.total_coffees >= 10,
    getProgress: (stats) => ({ current: stats.total_coffees, target: 10 }),
  },
  {
    id: 2,
    name: "Cascade",
    requirement: "Try 5 different origins",
    emoji: "💧",
    color: "#6890f0",
    checkUnlocked: (stats) => Object.keys(stats.origin_distribution).length >= 5,
    getProgress: (stats) => ({
      current: Object.keys(stats.origin_distribution).length,
      target: 5,
    }),
  },
  {
    id: 3,
    name: "Thunder",
    requirement: "Rate a coffee 10/10",
    emoji: "⚡",
    color: "#f8d030",
    checkUnlocked: (stats) => stats.highest_rated?.rating === 10,
    getProgress: (stats) => ({
      current: stats.highest_rated?.rating || 0,
      target: 10,
    }),
  },
  {
    id: 4,
    name: "Rainbow",
    requirement: "Collect 5 Pokemon types",
    emoji: "🌈",
    color: "#ff69b4",
    checkUnlocked: (stats) => Object.keys(stats.type_distribution).length >= 5,
    getProgress: (stats) => ({
      current: Object.keys(stats.type_distribution).length,
      target: 5,
    }),
  },
  {
    id: 5,
    name: "Soul",
    requirement: "Log 50 total brews",
    emoji: "💜",
    color: "#a040a0",
    checkUnlocked: (stats) => {
      const totalBrews = Object.values(stats.brewer_stats).reduce(
        (sum, b) => sum + b.count,
        0
      );
      return totalBrews >= 50;
    },
    getProgress: (stats) => {
      const totalBrews = Object.values(stats.brewer_stats).reduce(
        (sum, b) => sum + b.count,
        0
      );
      return { current: totalBrews, target: 50 };
    },
  },
  {
    id: 6,
    name: "Marsh",
    requirement: "Complete 25% of Pokedex",
    emoji: "🔮",
    color: "#9966cc",
    checkUnlocked: (stats) => stats.completion_percent >= 25,
    getProgress: (stats) => ({
      current: Math.round(stats.completion_percent),
      target: 25,
    }),
  },
  {
    id: 7,
    name: "Volcano",
    requirement: "Collect a Fire-type",
    emoji: "🔥",
    color: "#f08030",
    checkUnlocked: (stats) => (stats.type_distribution["Fire"] || 0) >= 1,
    getProgress: (stats) => ({
      current: stats.type_distribution["Fire"] || 0,
      target: 1,
    }),
  },
  {
    id: 8,
    name: "Earth",
    requirement: "Try all 5 processes",
    emoji: "🌍",
    color: "#78c850",
    checkUnlocked: (stats) => Object.keys(stats.processing_stats).length >= 5,
    getProgress: (stats) => ({
      current: Object.keys(stats.processing_stats).length,
      target: 5,
    }),
  },
];

interface GymBadgesProps {
  compact?: boolean;
}

const GymBadges: React.FC<GymBadgesProps> = ({ compact = false }) => {
  const [stats, setStats] = useState<StatisticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedBadge, setSelectedBadge] = useState<Badge | null>(null);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      const data = await api.getStatistics();
      setStats(data);
    } catch (error) {
      console.error("Failed to load stats for badges:", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="pokemon-textbox" style={{ textAlign: "center", padding: "12px" }}>
        <img
          src="./pokemon-sprites/animated/25.gif"
          alt="Loading"
          style={{ width: "48px", height: "auto", imageRendering: "pixelated" }}
        />
        <div style={{ fontSize: "9px", marginTop: "8px" }}>Loading badges...</div>
      </div>
    );
  }

  // Create empty stats if none loaded
  const safeStats: StatisticsData = stats || {
    total_coffees: 0,
    total_pokemon: 0,
    completion_percent: 0,
    average_rating: 0,
    highest_rated: null,
    type_distribution: {},
    origin_distribution: {},
    processing_stats: {},
    brewer_stats: {},
  };

  const unlockedCount = BADGES.filter((b) => b.checkUnlocked(safeStats)).length;

  if (compact) {
    // Compact view for home screen - just show badge icons
    return (
      <div className="pokemon-textbox">
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "8px",
          }}
        >
          <span style={{ fontWeight: "bold", fontSize: "10px" }}>GYM BADGES</span>
          <span style={{ fontSize: "9px" }}>{unlockedCount}/8</span>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(8, 1fr)",
            gap: "4px",
          }}
        >
          {BADGES.map((badge) => {
            const unlocked = badge.checkUnlocked(safeStats);
            return (
              <div
                key={badge.id}
                onClick={() => setSelectedBadge(badge)}
                style={{
                  width: "28px",
                  height: "28px",
                  borderRadius: "50%",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: "14px",
                  backgroundColor: unlocked ? badge.color : "#ccc",
                  opacity: unlocked ? 1 : 0.4,
                  cursor: "pointer",
                  border: "2px solid var(--gb-darkest)",
                  filter: unlocked ? "none" : "grayscale(100%)",
                  transition: "transform 0.1s",
                }}
                title={badge.name}
              >
                {unlocked ? badge.emoji : "?"}
              </div>
            );
          })}
        </div>

        {/* Badge detail popup */}
        {selectedBadge && (
          <div
            style={{
              marginTop: "8px",
              padding: "8px",
              backgroundColor: "var(--gb-lightest)",
              border: "2px solid var(--gb-darkest)",
              borderRadius: "4px",
              fontSize: "9px",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span style={{ fontWeight: "bold" }}>
                {selectedBadge.emoji} {selectedBadge.name} Badge
              </span>
              <button
                onClick={() => setSelectedBadge(null)}
                style={{
                  background: "none",
                  border: "none",
                  cursor: "pointer",
                  fontSize: "12px",
                }}
              >
                ✕
              </button>
            </div>
            <div style={{ marginTop: "4px" }}>{selectedBadge.requirement}</div>
            {(() => {
              const progress = selectedBadge.getProgress(safeStats);
              const unlocked = selectedBadge.checkUnlocked(safeStats);
              return (
                <div style={{ marginTop: "4px" }}>
                  {unlocked ? (
                    <span style={{ color: "#00aa00", fontWeight: "bold" }}>UNLOCKED!</span>
                  ) : (
                    <span>
                      Progress: {progress.current}/{progress.target}
                    </span>
                  )}
                </div>
              );
            })()}
          </div>
        )}
      </div>
    );
  }

  // Full view - could be used in a dedicated badges screen
  return (
    <div className="pokemon-textbox">
      <div style={{ fontWeight: "bold", fontSize: "12px", marginBottom: "12px", textAlign: "center" }}>
        KANTO GYM BADGES
      </div>
      <div style={{ textAlign: "center", fontSize: "10px", marginBottom: "12px" }}>
        {unlockedCount}/8 Badges Earned
      </div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(4, 1fr)",
          gap: "8px",
        }}
      >
        {BADGES.map((badge) => {
          const unlocked = badge.checkUnlocked(safeStats);
          const progress = badge.getProgress(safeStats);
          return (
            <div
              key={badge.id}
              style={{
                textAlign: "center",
                padding: "8px",
                backgroundColor: unlocked ? "var(--gb-light)" : "var(--gb-lightest)",
                border: "2px solid var(--gb-dark)",
                borderRadius: "4px",
                opacity: unlocked ? 1 : 0.6,
              }}
            >
              <div
                style={{
                  width: "36px",
                  height: "36px",
                  borderRadius: "50%",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: "18px",
                  backgroundColor: unlocked ? badge.color : "#ccc",
                  margin: "0 auto 4px",
                  border: "2px solid var(--gb-darkest)",
                  filter: unlocked ? "none" : "grayscale(100%)",
                }}
              >
                {unlocked ? badge.emoji : "?"}
              </div>
              <div style={{ fontSize: "8px", fontWeight: "bold" }}>{badge.name}</div>
              <div style={{ fontSize: "7px", marginTop: "2px" }}>
                {unlocked ? "EARNED!" : `${progress.current}/${progress.target}`}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default GymBadges;
