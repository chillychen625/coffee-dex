import React, { useState, useMemo, useCallback } from "react";
import { Brew, TastingTraits } from "../types/pokemon";

interface BrewChartProps {
  brews: Brew[];
}

type MetricKey = "rating" | keyof TastingTraits;

const METRIC_OPTIONS: { value: MetricKey; label: string }[] = [
  { value: "rating", label: "Rating" },
  { value: "sweetness", label: "Sweetness" },
  { value: "bitterness", label: "Bitterness" },
  { value: "body", label: "Body" },
  { value: "cleanliness", label: "Cleanliness" },
  { value: "berry_intensity", label: "Berry" },
  { value: "stonefruit_intensity", label: "Stonefruit" },
  { value: "citrus_fruits_intensity", label: "Citrus" },
  { value: "florality", label: "Florality" },
  { value: "aromatic_intensity", label: "Aromatic" },
  { value: "roast_intensity", label: "Roast" },
  { value: "spice", label: "Spice" },
  { value: "savory", label: "Savory" },
];

const CHART_COLOR = "#cc0000";
const GRID_COLOR = "#ccc";

// Chart layout constants
const SVG_WIDTH = 560;
const SVG_HEIGHT = 220;
const PADDING = { top: 20, right: 24, bottom: 40, left: 32 };
const PLOT_WIDTH = SVG_WIDTH - PADDING.left - PADDING.right;
const PLOT_HEIGHT = SVG_HEIGHT - PADDING.top - PADDING.bottom;

// Returns the metric value, or -1 if the trait was not scored
function getMetricValue(brew: Brew, metric: MetricKey): number {
  if (metric === "rating") {
    return brew.rating;
  }
  return brew.tasting_traits[metric];
}

export default function BrewChart({ brews }: BrewChartProps) {
  const [selectedMetric, setSelectedMetric] = useState<MetricKey>("rating");
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  const sortedBrews = useMemo(
    () =>
      [...brews].sort(
        (a, b) =>
          new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      ),
    [brews]
  );

  // Filter to only brews that have a scored value for this metric (skip -1)
  const scoredBrews = useMemo(
    () => sortedBrews.filter((brew) => getMetricValue(brew, selectedMetric) >= 0),
    [sortedBrews, selectedMetric]
  );

  const dataPoints = useMemo(
    () =>
      scoredBrews.map((brew, i) => {
        const value = getMetricValue(brew, selectedMetric);
        const x =
          scoredBrews.length === 1
            ? PADDING.left + PLOT_WIDTH / 2
            : PADDING.left + (i / (scoredBrews.length - 1)) * PLOT_WIDTH;
        const y = PADDING.top + PLOT_HEIGHT - (value / 10) * PLOT_HEIGHT;
        return { x, y, value, brew, index: i };
      }),
    [scoredBrews, selectedMetric]
  );

  const linePath = useMemo(() => {
    if (dataPoints.length < 2) return "";
    return dataPoints
      .map((p, i) => `${i === 0 ? "M" : "L"} ${p.x} ${p.y}`)
      .join(" ");
  }, [dataPoints]);

  const handleDotEnter = useCallback((index: number) => {
    setHoveredIndex(index);
  }, []);

  const handleDotLeave = useCallback(() => {
    setHoveredIndex(null);
  }, []);

  if (brews.length === 0) {
    return null;
  }

  const noData = scoredBrews.length === 0;

  // Y-axis labels
  const yLabels = [0, 5, 10];
  // Grid line at y=5
  const gridY5 = PADDING.top + PLOT_HEIGHT - (5 / 10) * PLOT_HEIGHT;

  return (
    <div className="pokemon-textbox" style={{ padding: "12px" }}>
      <div className="pokemon-subtitle" style={{ marginBottom: "8px" }}>
        BREW METRICS
      </div>

      <select
        className="pokemon-select"
        value={selectedMetric}
        onChange={(e) => setSelectedMetric(e.target.value as MetricKey)}
        style={{ marginBottom: "10px", width: "100%" }}
      >
        {METRIC_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>

      {noData ? (
        <div style={{ fontSize: "9px", opacity: 0.6, textAlign: "center", padding: "16px 0" }}>
          No brews scored this trait yet.
        </div>
      ) : (
      <div style={{ width: "100%", overflowX: "auto" }}>
        <svg
          viewBox={`0 0 ${SVG_WIDTH} ${SVG_HEIGHT}`}
          width="100%"
          style={{ maxWidth: `${SVG_WIDTH}px`, display: "block" }}
        >
          {/* Y-axis line */}
          <line
            x1={PADDING.left}
            y1={PADDING.top}
            x2={PADDING.left}
            y2={PADDING.top + PLOT_HEIGHT}
            stroke={GRID_COLOR}
            strokeWidth={1}
          />

          {/* X-axis line */}
          <line
            x1={PADDING.left}
            y1={PADDING.top + PLOT_HEIGHT}
            x2={PADDING.left + PLOT_WIDTH}
            y2={PADDING.top + PLOT_HEIGHT}
            stroke={GRID_COLOR}
            strokeWidth={1}
          />

          {/* Horizontal grid line at y=5 */}
          <line
            x1={PADDING.left}
            y1={gridY5}
            x2={PADDING.left + PLOT_WIDTH}
            y2={gridY5}
            stroke={GRID_COLOR}
            strokeWidth={1}
            strokeDasharray="4 4"
          />

          {/* Y-axis labels */}
          {yLabels.map((val) => {
            const yPos =
              PADDING.top + PLOT_HEIGHT - (val / 10) * PLOT_HEIGHT;
            return (
              <text
                key={val}
                x={PADDING.left - 6}
                y={yPos + 3}
                textAnchor="end"
                fontSize="10"
                fill="#555"
                fontFamily="monospace"
              >
                {val}
              </text>
            );
          })}

          {/* Line path */}
          {dataPoints.length >= 2 && (
            <path
              d={linePath}
              fill="none"
              stroke={CHART_COLOR}
              strokeWidth={2}
              strokeLinejoin="round"
              strokeLinecap="round"
            />
          )}

          {/* Data points and labels */}
          {dataPoints.map((point) => (
            <g key={point.index}>
              {/* Days off roast label on X-axis */}
              {point.brew.days_off_roast >= 0 && (
                <text
                  x={point.x}
                  y={PADDING.top + PLOT_HEIGHT + 14}
                  textAnchor="middle"
                  fontSize="9"
                  fill="#777"
                  fontFamily="monospace"
                >
                  d{point.brew.days_off_roast}
                </text>
              )}

              {/* Brew number label */}
              <text
                x={point.x}
                y={PADDING.top + PLOT_HEIGHT + 28}
                textAnchor="middle"
                fontSize="8"
                fill="#aaa"
                fontFamily="monospace"
              >
                #{point.index + 1}
              </text>

              {/* Hover tooltip */}
              {hoveredIndex === point.index && (
                <g>
                  <rect
                    x={point.x - 16}
                    y={point.y - 22}
                    width={32}
                    height={16}
                    rx={3}
                    fill="#333"
                  />
                  <text
                    x={point.x}
                    y={point.y - 10}
                    textAnchor="middle"
                    fontSize="10"
                    fontWeight="bold"
                    fill="#fff"
                    fontFamily="monospace"
                  >
                    {point.value}
                  </text>
                </g>
              )}

              {/* Dot */}
              <circle
                cx={point.x}
                cy={point.y}
                r={hoveredIndex === point.index ? 6 : 4}
                fill={CHART_COLOR}
                stroke="#fff"
                strokeWidth={2}
                style={{ cursor: "pointer", transition: "r 0.15s ease" }}
                onMouseEnter={() => handleDotEnter(point.index)}
                onMouseLeave={handleDotLeave}
              />
            </g>
          ))}
        </svg>
      </div>
      )}
    </div>
  );
}
