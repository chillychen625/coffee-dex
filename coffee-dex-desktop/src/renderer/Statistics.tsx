import React, { useState, useEffect } from "react";
import { api } from "../services/api";
import { Brew } from "../types/pokemon";
import LoadingSpinner from "./LoadingSpinner";
import EmptyState from "./EmptyState";
import FormDecoSprites from "./FormDecoSprites";

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

// Pokemon type colors for visual bars
const typeColors: { [key: string]: string } = {
  Fire: "#F08030",
  Water: "#6890F0",
  Grass: "#78C850",
  Electric: "#F8D030",
  Psychic: "#F85888",
  Ice: "#98D8D8",
  Dragon: "#7038F8",
  Dark: "#705848",
  Fairy: "#EE99AC",
  Normal: "#A8A878",
  Fighting: "#C03028",
  Flying: "#A890F0",
  Poison: "#A040A0",
  Ground: "#E0C068",
  Rock: "#B8A038",
  Bug: "#A8B820",
  Ghost: "#705898",
  Steel: "#B8B8D0",
};

// Bar component for visual stats
const StatBar: React.FC<{
  label: string;
  value: number;
  maxValue: number;
  color?: string;
  showValue?: boolean;
}> = ({ label, value, maxValue, color = "#4a9eff", showValue = true }) => {
  const percentage = Math.min((value / maxValue) * 100, 100);
  return (
    <div style={{ marginBottom: "4px" }}>
      <div style={{ display: "flex", justifyContent: "space-between", fontSize: "7px", marginBottom: "2px" }}>
        <span style={{ textTransform: "capitalize" }}>{label.replace(/_/g, " ")}</span>
        {showValue && <span style={{ fontWeight: "bold" }}>{typeof value === 'number' ? value.toFixed(1) : value}</span>}
      </div>
      <div style={{
        height: "6px",
        background: "rgba(0,0,0,0.2)",
        borderRadius: "3px",
        overflow: "hidden"
      }}>
        <div style={{
          height: "100%",
          width: `${percentage}%`,
          background: color,
          borderRadius: "3px",
          transition: "width 0.3s ease"
        }} />
      </div>
    </div>
  );
};

// Circular progress component
const CircularProgress: React.FC<{
  value: number;
  maxValue: number;
  label: string;
  size?: number;
  color?: string;
}> = ({ value, maxValue, label, size = 60, color = "#4a9eff" }) => {
  const percentage = (value / maxValue) * 100;
  const strokeWidth = 6;
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (percentage / 100) * circumference;

  return (
    <div style={{ textAlign: "center" }}>
      <svg width={size} height={size} style={{ transform: "rotate(-90deg)" }}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="rgba(0,0,0,0.2)"
          strokeWidth={strokeWidth}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          strokeLinecap="round"
          style={{ transition: "stroke-dashoffset 0.5s ease" }}
        />
      </svg>
      <div style={{ marginTop: "-45px", fontSize: "12px", fontWeight: "bold" }}>
        {percentage.toFixed(0)}%
      </div>
      <div style={{ fontSize: "7px", marginTop: "2px" }}>{label}</div>
    </div>
  );
};

// Helper to group brews by day
interface DayBrews {
  date: string;
  dayLabel: string;
  brews: Brew[];
  totalRating: number;
  avgRating: number;
}

const groupBrewsByDay = (brews: Brew[]): DayBrews[] => {
  const now = new Date();
  const fiveDaysAgo = new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000);

  // Filter to last 5 days
  const recentBrews = brews.filter(b => {
    if (!b.created_at) return false;
    const brewDate = new Date(b.created_at);
    return brewDate >= fiveDaysAgo;
  });

  // Group by date
  const groups: { [key: string]: Brew[] } = {};
  recentBrews.forEach(brew => {
    const date = new Date(brew.created_at!).toISOString().split('T')[0];
    if (!groups[date]) groups[date] = [];
    groups[date].push(brew);
  });

  // Convert to array and add labels
  const dayLabels = ['Today', 'Yesterday'];
  return Object.entries(groups)
    .sort(([a], [b]) => b.localeCompare(a)) // Sort newest first
    .map(([date, dayBrews]) => {
      const brewDate = new Date(date);
      const diffDays = Math.floor((now.getTime() - brewDate.getTime()) / (1000 * 60 * 60 * 24));
      const dayLabel = diffDays < dayLabels.length ? dayLabels[diffDays] : brewDate.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
      const totalRating = dayBrews.reduce((sum, b) => sum + (b.rating || 0), 0);
      return {
        date,
        dayLabel,
        brews: dayBrews,
        totalRating,
        avgRating: dayBrews.length > 0 ? totalRating / dayBrews.length : 0
      };
    });
};

const Statistics: React.FC<StatisticsProps> = ({ onBack }) => {
  const [stats, setStats] = useState<StatisticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [aiInsight, setAiInsight] = useState<string | null>(null);
  const [aiLoading, setAiLoading] = useState(false);
  const [showApiKeyInput, setShowApiKeyInput] = useState(false);
  const [apiKey, setApiKey] = useState("");
  const [recentBrews, setRecentBrews] = useState<Brew[]>([]);
  const [showRecentActivity, setShowRecentActivity] = useState(false);
  const [recentLoading, setRecentLoading] = useState(false);

  useEffect(() => {
    loadStatistics();
  }, []);

  const loadRecentActivity = async () => {
    if (recentBrews.length > 0) {
      // Already loaded, just toggle
      setShowRecentActivity(!showRecentActivity);
      return;
    }
    setRecentLoading(true);
    try {
      const brews = await api.getRecentBrews();
      setRecentBrews(brews || []);
      setShowRecentActivity(true);
    } catch (err) {
      console.error("Failed to load recent brews:", err);
    } finally {
      setRecentLoading(false);
    }
  };

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

  const generateAIInsight = async () => {
    if (!stats) return;

    // Check for API key: env first, then localStorage, then manual input
    let storedKey = apiKey || localStorage.getItem("anthropic_api_key") || "";

    // Try to get from environment via Electron IPC
    if (!storedKey && window.electron?.getAnthropicApiKey) {
      try {
        const envKey = await window.electron.getAnthropicApiKey();
        if (envKey) {
          storedKey = envKey;
        }
      } catch (e) {
        console.log("Could not get API key from env:", e);
      }
    }

    if (!storedKey) {
      setShowApiKeyInput(true);
      return;
    }

    setAiLoading(true);
    setAiInsight(null);

    try {
      // Build a comprehensive prompt with all the statistics
      const prompt = `You are a fun, knowledgeable coffee expert and Pokemon trainer analyzing a coffee collection. Based on these statistics, provide a brief, engaging analysis (3-4 sentences) with personality. Include a coffee recommendation or insight about their taste preferences.

Collection Stats:
- Total Coffees: ${stats.total_coffees}
- Total Pokemon Collected: ${stats.total_pokemon} (${stats.completion_percent.toFixed(1)}% of Gen 1)
- Average Rating: ${stats.average_rating.toFixed(1)}/10
- Best Coffee: ${stats.highest_rated?.name || "N/A"} (${stats.highest_rated?.rating || 0}/10)
- Worst Coffee: ${stats.lowest_rated?.name || "N/A"} (${stats.lowest_rated?.rating || 0}/10)
- Top Pokemon Types: ${Object.entries(stats.type_distribution).sort(([,a], [,b]) => b - a).slice(0, 3).map(([t, c]) => `${t}(${c})`).join(", ")}
- Top Origins: ${stats.top_origins.slice(0, 3).map(o => `${o.origin}(${o.count})`).join(", ")}
- Flavor Profile Highlights: ${Object.entries(stats.trait_averages).filter(([_, v]) => v >= 6).map(([t, v]) => `${t.replace(/_/g, " ")}(${v.toFixed(1)})`).join(", ") || "Balanced across all traits"}
- Favorite Processing: ${Object.entries(stats.processing_stats).sort(([,a], [,b]) => b.count - a.count)[0]?.[0] || "Various"}
- Mapping Confidence: ${(stats.average_confidence * 100).toFixed(1)}%

Be creative but concise! Reference Pokemon types and coffee characteristics.`;

      const response = await fetch("https://api.anthropic.com/v1/messages", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-api-key": storedKey,
          "anthropic-version": "2023-06-01",
          "anthropic-dangerous-direct-browser-access": "true"
        },
        body: JSON.stringify({
          model: "claude-3-5-haiku-20241022",
          max_tokens: 300,
          messages: [{ role: "user", content: prompt }]
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error?.message || `API error: ${response.status}`);
      }

      const data = await response.json();
      const insight = data.content?.[0]?.text || "Unable to generate insight";
      setAiInsight(insight);

      // Store the API key for future use
      if (apiKey) {
        localStorage.setItem("anthropic_api_key", apiKey);
      }
    } catch (err) {
      console.error("AI insight error:", err);
      setAiInsight(`Failed to generate insight: ${err}`);
    } finally {
      setAiLoading(false);
    }
  };

  const handleApiKeySubmit = () => {
    if (apiKey.trim()) {
      localStorage.setItem("anthropic_api_key", apiKey);
      setShowApiKeyInput(false);
      generateAIInsight();
    }
  };

  if (loading) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ position: "relative" }}
        >
          <FormDecoSprites seed="stats-loading" spin={true} />
          <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 1 }}>
            ← Back
          </button>
          <LoadingSpinner variant="default" message="Loading Statistics..." />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ position: "relative" }}
        >
          <FormDecoSprites seed="stats-error" spin={true} />
          <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 1 }}>
            ← Back
          </button>
          <h2 className="pokemon-title" style={{ fontSize: "14px", position: "relative", zIndex: 1 }}>
            STATISTICS ERROR
          </h2>
          <EmptyState
            variant="error"
            title="Failed to Load"
            message={error}
          >
            <button className="pokemon-button" onClick={loadStatistics}>
              Retry
            </button>
          </EmptyState>
        </div>
      </div>
    );
  }

  if (!stats || stats.total_coffees === 0) {
    return (
      <div className="pokemon-screen centered">
        <div
          className="pokemon-frame"
          style={{ position: "relative" }}
        >
          <FormDecoSprites seed="stats-empty" spin={true} />
          <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 1 }}>
            ← Back
          </button>
          <h2 className="pokemon-title" style={{ fontSize: "14px", position: "relative", zIndex: 1 }}>
            STATISTICS
          </h2>
          <EmptyState variant="no-stats" />
        </div>
      </div>
    );
  }

  const maxTypeCount = Math.max(...Object.values(stats.type_distribution), 1);

  return (
    <div className="pokemon-screen">
      <div
        className="pokemon-frame"
        style={{ position: "relative", paddingBottom: "24px" }}
      >
        <FormDecoSprites seed="statistics" spin={true} />
        <div style={{
          position: "sticky",
          top: 0,
          zIndex: 10,
          background: "inherit",
          paddingTop: "4px",
          paddingBottom: "4px",
        }}>
          <button className="pokemon-button mb-md" onClick={onBack} style={{ position: "relative", zIndex: 11 }}>
            ← Back
          </button>
        </div>

        <h2 className="pokemon-title" style={{ fontSize: "14px", position: "relative", zIndex: 1 }}>
          STATISTICS
        </h2>

        {/* Visual Overview with Circular Progress */}
        <div className="pokemon-textbox mb-md" style={{ fontSize: "10px" }}>
          <div style={{ fontWeight: "bold", marginBottom: "12px", textAlign: "center" }}>
            COLLECTION OVERVIEW
          </div>
          <div style={{ display: "flex", justifyContent: "space-around", marginBottom: "12px" }}>
            <CircularProgress
              value={stats.completion_percent}
              maxValue={100}
              label="Pokedex"
              color="#78C850"
            />
            <CircularProgress
              value={stats.average_rating * 10}
              maxValue={100}
              label="Avg Rating"
              color="#F8D030"
            />
            <CircularProgress
              value={stats.average_confidence * 100}
              maxValue={100}
              label="Confidence"
              color="#6890F0"
            />
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px", fontSize: "9px" }}>
            <div><strong>Total Coffees:</strong> {stats.total_coffees}</div>
            <div><strong>Pokemon:</strong> {stats.total_pokemon}/151</div>
          </div>
          <button
            className="pokemon-button mt-md"
            onClick={loadRecentActivity}
            style={{ width: "100%", fontSize: "9px", padding: "8px" }}
            disabled={recentLoading}
          >
            {recentLoading ? "Loading..." : showRecentActivity ? "Hide Recent Activity" : "Show 5-Day Activity"}
          </button>
        </div>

        {/* Recent Activity (Past 5 Days) */}
        {showRecentActivity && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              RECENT ACTIVITY (PAST 5 DAYS)
            </div>
            {(() => {
              const grouped = groupBrewsByDay(recentBrews);
              if (grouped.length === 0) {
                return <div style={{ textAlign: "center", opacity: 0.7 }}>No brews in the past 5 days</div>;
              }
              return grouped.map(day => (
                <div key={day.date} style={{ marginBottom: "12px" }}>
                  <div style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginBottom: "4px",
                    padding: "4px 8px",
                    background: "rgba(0,0,0,0.1)",
                    borderRadius: "4px"
                  }}>
                    <span style={{ fontWeight: "bold" }}>{day.dayLabel}</span>
                    <span>{day.brews.length} brew{day.brews.length !== 1 ? 's' : ''} • avg {day.avgRating.toFixed(1)}/10</span>
                  </div>
                  {day.brews.map(brew => (
                    <div key={brew.id} style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "8px",
                      padding: "4px 8px",
                      marginLeft: "8px",
                      borderLeft: "2px solid rgba(0,0,0,0.2)"
                    }}>
                      <span style={{
                        width: "24px",
                        height: "24px",
                        borderRadius: "50%",
                        background: `hsl(${(brew.rating || 5) * 12}, 70%, 50%)`,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        color: "white",
                        fontWeight: "bold",
                        fontSize: "10px"
                      }}>
                        {brew.rating}
                      </span>
                      <div style={{ flex: 1 }}>
                        <div style={{ fontWeight: "bold", fontSize: "8px" }}>
                          {brew.dripper || "Unknown brewer"}
                        </div>
                        {brew.tasting_notes && brew.tasting_notes.filter(n => n).length > 0 && (
                          <div style={{ fontSize: "7px", opacity: 0.8 }}>
                            {brew.tasting_notes.filter(n => n).slice(0, 3).join(", ")}
                          </div>
                        )}
                      </div>
                      {brew.end_time && (
                        <div style={{ fontSize: "8px", opacity: 0.7 }}>
                          {brew.end_time.minutes}:{String(brew.end_time.seconds).padStart(2, "0")}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ));
            })()}
          </div>
        )}

        {/* Rating Extremes with Visual */}
        {(stats.highest_rated || stats.lowest_rated) && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              RATING EXTREMES
            </div>
            {stats.highest_rated && (
              <div style={{ marginBottom: "8px" }}>
                <div style={{ display: "flex", alignItems: "center", gap: "8px", marginBottom: "4px" }}>
                  <span>⭐</span>
                  <span style={{ flex: 1 }}>{stats.highest_rated.name}</span>
                  <span style={{ fontWeight: "bold", color: "#78C850" }}>{stats.highest_rated.rating}/10</span>
                </div>
                <div style={{
                  height: "4px",
                  background: "rgba(0,0,0,0.2)",
                  borderRadius: "2px",
                  overflow: "hidden"
                }}>
                  <div style={{
                    height: "100%",
                    width: `${stats.highest_rated.rating * 10}%`,
                    background: "linear-gradient(90deg, #78C850, #98D850)",
                    borderRadius: "2px"
                  }} />
                </div>
              </div>
            )}
            {stats.lowest_rated && (
              <div>
                <div style={{ display: "flex", alignItems: "center", gap: "8px", marginBottom: "4px" }}>
                  <span>⚠️</span>
                  <span style={{ flex: 1 }}>{stats.lowest_rated.name}</span>
                  <span style={{ fontWeight: "bold", color: "#F08030" }}>{stats.lowest_rated.rating}/10</span>
                </div>
                <div style={{
                  height: "4px",
                  background: "rgba(0,0,0,0.2)",
                  borderRadius: "2px",
                  overflow: "hidden"
                }}>
                  <div style={{
                    height: "100%",
                    width: `${stats.lowest_rated.rating * 10}%`,
                    background: "linear-gradient(90deg, #F08030, #F0A030)",
                    borderRadius: "2px"
                  }} />
                </div>
              </div>
            )}
          </div>
        )}

        {/* Type Distribution with Visual Bars */}
        {Object.keys(stats.type_distribution).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              POKEMON TYPE DISTRIBUTION
            </div>
            {Object.entries(stats.type_distribution)
              .sort(([, a], [, b]) => b - a)
              .slice(0, 6)
              .map(([type, count]) => (
                <StatBar
                  key={type}
                  label={type}
                  value={count}
                  maxValue={maxTypeCount}
                  color={typeColors[type] || "#A8A878"}
                  showValue={true}
                />
              ))}
          </div>
        )}

        {/* Flavor Profile with Visual Bars */}
        {Object.keys(stats.trait_averages).length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "8px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              AVERAGE FLAVOR PROFILE
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px" }}>
              <div>
                <StatBar label="Berry" value={stats.trait_averages.berry_intensity} maxValue={10} color="#E0457B" />
                <StatBar label="Stonefruit" value={stats.trait_averages.stonefruit_intensity} maxValue={10} color="#FFB347" />
                <StatBar label="Citrus" value={stats.trait_averages.citrus_fruits_intensity} maxValue={10} color="#FDD835" />
                <StatBar label="Floral" value={stats.trait_averages.florality} maxValue={10} color="#BA68C8" />
                <StatBar label="Sweetness" value={stats.trait_averages.sweetness} maxValue={10} color="#FF8A80" />
                <StatBar label="Aroma" value={stats.trait_averages.aromatic_intensity} maxValue={10} color="#80DEEA" />
              </div>
              <div>
                <StatBar label="Roast" value={stats.trait_averages.roast_intensity} maxValue={10} color="#795548" />
                <StatBar label="Bitterness" value={stats.trait_averages.bitterness} maxValue={10} color="#5D4037" />
                <StatBar label="Spice" value={stats.trait_averages.spice} maxValue={10} color="#FF5722" />
                <StatBar label="Savory" value={stats.trait_averages.savory} maxValue={10} color="#827717" />
                <StatBar label="Body" value={stats.trait_averages.body} maxValue={10} color="#6D4C41" />
                <StatBar label="Cleanliness" value={stats.trait_averages.cleanliness} maxValue={10} color="#4FC3F7" />
              </div>
            </div>
          </div>
        )}

        {/* Top Origins with Bars */}
        {stats.top_origins.length > 0 && (
          <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
            <div style={{ fontWeight: "bold", marginBottom: "8px" }}>
              TOP ORIGINS
            </div>
            {stats.top_origins.slice(0, 5).map((origin) => (
              <div key={origin.origin} style={{ marginBottom: "6px" }}>
                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "2px" }}>
                  <span>{origin.origin}</span>
                  <span style={{ fontSize: "8px" }}>{origin.count} coffees • avg {origin.average_rating.toFixed(1)}/10</span>
                </div>
                <div style={{
                  height: "4px",
                  background: "rgba(0,0,0,0.2)",
                  borderRadius: "2px",
                  overflow: "hidden"
                }}>
                  <div style={{
                    height: "100%",
                    width: `${origin.average_rating * 10}%`,
                    background: `hsl(${origin.average_rating * 12}, 70%, 50%)`,
                    borderRadius: "2px"
                  }} />
                </div>
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
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "4px" }}>
              {Object.entries(stats.processing_stats).map(([method, stat]) => (
                <div key={method} style={{
                  textTransform: "capitalize",
                  padding: "4px 8px",
                  background: "rgba(0,0,0,0.1)",
                  borderRadius: "4px",
                  fontSize: "8px"
                }}>
                  <div style={{ fontWeight: "bold" }}>{method}</div>
                  <div>{stat.count} coffees • {stat.average_rating.toFixed(1)}/10</div>
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
            <div style={{ display: "flex", gap: "4px", flexWrap: "wrap" }}>
              {Object.entries(stats.roast_distribution).map(([level, count]) => (
                <div key={level} style={{
                  textTransform: "capitalize",
                  padding: "4px 8px",
                  background: level.includes("dark") ? "#5D4037" : level.includes("light") ? "#BCAAA4" : "#8D6E63",
                  color: "white",
                  borderRadius: "4px",
                  fontSize: "8px"
                }}>
                  {level}: {count}
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

        {/* AI Insight Section */}
        <div className="pokemon-textbox mb-md" style={{ fontSize: "9px" }}>
          <div style={{ fontWeight: "bold", marginBottom: "8px", display: "flex", alignItems: "center", gap: "8px" }}>
            <span>🤖</span> AI COFFEE INSIGHT
          </div>

          {showApiKeyInput ? (
            <div>
              <div style={{ marginBottom: "8px", fontSize: "8px" }}>
                Enter your Anthropic API key to enable AI insights:
              </div>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk-ant-..."
                style={{
                  width: "100%",
                  padding: "6px",
                  marginBottom: "8px",
                  border: "1px solid #000",
                  borderRadius: "4px",
                  fontSize: "9px"
                }}
              />
              <div style={{ display: "flex", gap: "8px" }}>
                <button
                  className="pokemon-button"
                  onClick={handleApiKeySubmit}
                  style={{ flex: 1, fontSize: "9px", padding: "6px" }}
                >
                  Save & Generate
                </button>
                <button
                  className="pokemon-button"
                  onClick={() => setShowApiKeyInput(false)}
                  style={{ fontSize: "9px", padding: "6px" }}
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : aiLoading ? (
            <div style={{ textAlign: "center", padding: "12px" }}>
              <div style={{ marginBottom: "8px" }}>Generating insight...</div>
              <div className="pokemon-loading-dots">
                <span>●</span><span>●</span><span>●</span>
              </div>
            </div>
          ) : aiInsight ? (
            <div>
              <div style={{ lineHeight: "1.5", marginBottom: "8px" }}>
                {aiInsight}
              </div>
              <button
                className="pokemon-button"
                onClick={generateAIInsight}
                style={{ fontSize: "8px", padding: "4px 8px" }}
              >
                Regenerate
              </button>
            </div>
          ) : (
            <div>
              <div style={{ marginBottom: "8px", fontSize: "8px", opacity: 0.8 }}>
                Get a personalized AI analysis of your coffee collection and taste preferences.
              </div>
              <button
                className="pokemon-button"
                onClick={generateAIInsight}
                style={{ width: "100%", fontSize: "9px", padding: "8px" }}
              >
                Generate AI Insight
              </button>
            </div>
          )}
        </div>

        {/* Confidence Metrics */}
        <div className="pokemon-textbox" style={{ fontSize: "9px" }}>
          <div style={{ fontWeight: "bold", marginBottom: "4px" }}>
            POKEMON MAPPING CONFIDENCE
          </div>
          <div style={{ marginBottom: "8px" }}>
            <StatBar
              label="Average Confidence"
              value={stats.average_confidence * 100}
              maxValue={100}
              color={stats.average_confidence > 0.7 ? "#78C850" : stats.average_confidence > 0.4 ? "#F8D030" : "#F08030"}
            />
          </div>
          <div style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: "4px",
            fontSize: "8px",
          }}>
            <div style={{
              padding: "4px",
              background: "rgba(120, 200, 80, 0.2)",
              borderRadius: "4px"
            }}>
              High (&gt;80%): {stats.high_confidence_pairings}
            </div>
            <div style={{
              padding: "4px",
              background: "rgba(104, 144, 240, 0.2)",
              borderRadius: "4px"
            }}>
              Total Pokemon: {stats.total_pokemon}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Statistics;
