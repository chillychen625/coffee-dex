const esbuild = require("esbuild");
const path = require("path");

// Build renderer process
esbuild
  .build({
    entryPoints: ["src/renderer/index.tsx"],
    bundle: true,
    outfile: "dist/renderer.js",
    platform: "browser",
    target: "es2020",
    loader: {
      ".tsx": "tsx",
      ".ts": "ts",
      ".css": "css",
      ".woff": "file",
      ".woff2": "file",
      ".ttf": "file",
      ".eot": "file",
    },
    external: ["electron"],
    minify: process.env.NODE_ENV === "production",
    sourcemap: process.env.NODE_ENV !== "production",
    assetNames: "assets/[name]-[hash]",
  })
  .then(() => {
    console.log("✓ Renderer process built successfully");
  })
  .catch((error) => {
    console.error("✗ Build failed:", error);
    process.exit(1);
  });
