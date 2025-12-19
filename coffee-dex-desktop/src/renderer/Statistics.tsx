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
    origin?: string;
    pokemon_name?: string;
  } | null;
  lowest_rated: {
    name: string;
    rating: number;
    origin?: string;
    pokemon_name?: string;
  } | null;
  type_distribution: { [key: string]: number };
  most_common_type: string;
  origin_distribution: { [key: string]: number };
  top_origins: Array<{ origin: string; count: number; average_rating: number }>;
  processing_stats: {
    [key: string]: {
      count: number;
      average_rating: number;
      common_types: string[];
    };
  };
  roast_distribution: { [key: string]: number };
  trait_averages: {
    berry_intensity: number;
    stonefruit_intensity: number;
    roast_intensity: number;
    citrus_fruits_intensity: number;
    bitterness: number;
    florality: number;
    spice: number;
    sweetness: number;
    aromatic_intensity: number;
    savory: number;
    body: number;
    cleanliness: number;
  };
  trait_ranges: {
    berry_range: { min: number; max: number };
    stonefruit_range: { min: number; max: number };
    roast_range: { min: number; max: number };
    citrus_range: { min: number; max: number };
    bitterness_range: { min: number; max: number };
    florality_range: { min: number; max: number };
    spice_range: { min: number; max: number };
    sweetness_range: { min: number; max: number };
    aromatic_range: { min: number; max: number };
    savory_range: { min: number; max: number };
    body_range: { min: number; max: number };
    cleanliness_range: { min: number; max: number };
  };
  brewer_stats: {
    [key: string]: {
      count: number;
      average_rating: number;
      avg_brew_time_seconds: number;
    };
  };
  average_confidence: number;
  high_confidence_pairings: number;
}

interface StatisticsProps {
  onBack: () => void;
}

const Statistics: React.FC<StatisticsProps> = ({ onBack }) => {
  const [stats, setStats] = useState<StatisticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadStatistics();
  }, []);

  const loadStatistics = async () => {
    setLoading(true);
    setError(null);
    try {
      console.log("Loading statistics...");
      const data = await api.getStatistics();
      console.log("Statistics data received:", data);
      setStats(data);
    } catch (err) {
      console.error("Statistics error:", err);
      setError(`Failed to load statistics: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  console.log(
    "Statistics render - loading:",
    loading,
    "error:",
    error,
    "stats:",
    stats
  );

  if (loading) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button className="pokemon-button mb-md" onClick={onBack}>
            ← Back
          </button>
          <div className="pokemon-loading">Loading Statistics</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button className="pokemon-button mb-md" onClick={onBack}>
            ← Back
          </button>
          <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
            STATISTICS ERROR
          </h2>
          <div
            className="pokemon-textbox"
            style={{ color: "#cc0000", fontSize: "10px" }}
          >
            {error}
          </div>
          <button className="pokemon-button mt-md" onClick={loadStatistics}>
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button className="pokemon-button mb-md" onClick={onBack}>
            ← Back
          </button>
          <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
            STATISTICS
          </h2>
          <div className="pokemon-textbox" style={{ textAlign: "center" }}>
            <div style={{ fontSize: "24px", marginBottom: "8px" }}>📊</div>
            <div>No statistics available yet!</div>
            <div style={{ fontSize: "8px", marginTop: "8px" }}>
              Create some coffee entries to see your stats.
            </div>
          </div>
        </div>
      </div>
    );
  }

  // If we have stats but everything is empty/zero
  if (stats.total_coffees === 0) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ maxWidth: "600px", margin: "0 auto" }}
        >
          <button className="pokemon-button mb-md" onClick={onBack}>
            ← Back
          </button>
          <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
            STATISTICS
          </h2>
          <div className="pokemon-textbox" style={{ textAlign: "center" }}>
            <div style={{ fontSize: "24px", marginBottom: "8px" }}>📊</div>
            <div>No coffee entries yet!</div>
            <div style={{ fontSize: "8px", marginTop: "8px" }}>
              Start brewing and tracking to build your coffee statistics.
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="pokemon-screen centered">
      <div
        className="pokemon-frame"
        style={{ maxWidth: "600px", margin: "0 auto" }}
      >
        <button className="pokemon-button mb-md" onClick={onBack}>
          ← Back
        </button>

        <h2 className="pokemon-title" style={{ fontSize: "14px" }}>
          STATISTICS
        </h2>

        {/* Overview */}
        <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
          <div
            style={{
              fontWeight: "bold",
              marginBottom: "8px",
              textAlign: "center",
            }}
          >
            COLLECTION OVERVIEW
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "8px",
            }}
          >
            <div>
              <strong>Total Coffees:</strong> {stats.total_coffees}
            </div>
            <div>
              <strong>Unique Pokemon:</strong> {stats.total_pokemon}
            </div>
            <div>
              <strong>Completion:</strong> {stats.completion_percent.toFixed(1)}
              %
            </div>
            <div>
              <strong>Avg Rating:</strong> {stats.average_rating.toFixed(1)}/10
            </div>
          </div>
        </div>

        {/* Rating Extremes */}
        {(stats.highest_rated || stats.lowest_rated) && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              RATING EXTREMES
            </div>
            {stats.highest_rated && (
              <div>
                ⭐ Best: {stats.highest_rated.name} (
                {stats.highest_rated.rating}/10)
              </div>
            )}
            {stats.lowest_rated && (
              <div>
                ⚠️ Worst: {stats.lowest_rated.name} ({stats.lowest_rated.rating}
                /10)
              </div>
            )}
          </div>
        )}

        {/* Type Distribution */}
        {Object.keys(stats.type_distribution).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              TYPE DISTRIBUTION
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "4px",
              }}
            >
              {Object.entries(stats.type_distribution)
                .sort(([, a], [, b]) => b - a)
                .slice(0, 8)
                .map(([type, count]) => (
                  <div key={type}>
                    {type}: {count}
                  </div>
                ))}
            </div>
          </div>
        )}

        {/* Top Origins */}
        {stats.top_origins.length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              TOP ORIGINS
            </div>
            {stats.top_origins.slice(0, 5).map((origin) => (
              <div key={origin.origin} style={{ marginBottom: "2px" }}>
                ▸ {origin.origin}: {origin.count} coffees (avg{" "}
                {origin.average_rating.toFixed(1)}/10)
              </div>
            ))}
          </div>
        )}

        {/* Processing Methods */}
        {Object.keys(stats.processing_stats).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              PROCESSING METHODS
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "4px",
              }}
            >
              {Object.entries(stats.processing_stats).map(([method, stat]) => (
                <div key={method} style={{ textTransform: "capitalize" }}>
                  {method}: {stat.count}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Roast Levels */}
        {Object.keys(stats.roast_distribution).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              ROAST LEVELS
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "4px",
              }}
            >
              {Object.entries(stats.roast_distribution).map(
                ([level, count]) => (
                  <div key={level} style={{ textTransform: "capitalize" }}>
                    {level}: {count}
                  </div>
                )
              )}
            </div>
          </div>
        )}

        {/* Trait Averages */}
        {Object.keys(stats.trait_averages).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "8px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              AVERAGE FLAVOR PROFILE
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "4px",
              }}
            >
              {Object.entries(stats.trait_averages).map(([trait, avg]) => (
                <div
                  key={trait}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <div style={{ textTransform: "capitalize", fontSize: "7px" }}>
                    {trait.replace(/_/g, " ")}
                  </div>
                  <div style={{ fontWeight: "bold" }}>{avg.toFixed(1)}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Brewer Stats */}
        {Object.keys(stats.brewer_stats).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
              BREWER STATISTICS
            </div>
            {Object.entries(stats.brewer_stats)
              .slice(0, 5)
              .map(([brewer, stat]) => (
                <div key={brewer} style={{ marginBottom: "4px" }}>
                  <div>
                    ▸ {brewer}: {stat.count} brews
                  </div>
                  <div style={{ fontSize: "8px", marginLeft: "8px" }}>
                    Avg: {stat.average_rating.toFixed(1)}/10 •{" "}
                    {Math.floor(stat.avg_brew_time_seconds / 60)}:
                    {String(
                      Math.floor(stat.avg_brew_time_seconds % 60)
                    ).padStart(2, "0")}
                  </div>
                </div>
              ))}
          </div>
        )}

        {/* Confidence Metrics */}
        <div className="pokemon-textbox" style={{ fontSize: "9px" }}>
          <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
            POKEMON MAPPING CONFIDENCE
          </div>
          <div>Average: {(stats.average_confidence * 100).toFixed(1)}%</div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "4px",
              marginTop: "4px",
              fontSize: "8px",
            }}
          >
            <div>High (&gt;80%): {stats.high_confidence_pairings}</div>
            <div>Total Pokemon: {stats.total_pokemon}</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Statistics;
