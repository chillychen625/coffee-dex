const esbuild = require("esbuild");
const path = require("path");
const fs = require("fs");

// Ensure dist directory exists
const distDir = path.join(__dirname, "dist");
if (!fs.existsSync(distDir)) {
  fs.mkdirSync(distDir, { recursive: true });
}

// Copy index.html to dist
const indexSrc = path.join(__dirname, "static", "index.html");
const indexDest = path.join(distDir, "index.html");
fs.copyFileSync(indexSrc, indexDest);
console.log("✓ Copied index.html to dist/");

// Ensure dist/pokemon-sprites directory exists
const spritesDir = path.join(__dirname, "dist", "pokemon-sprites");
if (!fs.existsSync(spritesDir)) {
  fs.mkdirSync(spritesDir, { recursive: true });
}

// Copy pokeball sprite files
const pokeballSprites = [
  "fast-ball.png",
  "lagreat-ball.png",
  "laultra-ball.png",
  "left-poke-ball.png",
];

pokeballSprites.forEach((sprite) => {
  const src = path.join(__dirname, "static", "pokemon-sprites", sprite);
  const dest = path.join(spritesDir, sprite);
  if (fs.existsSync(src)) {
    fs.copyFileSync(src, dest);
    console.log(`✓ Copied ${sprite}`);
  }
});

// Build renderer process
esbuild
  .build({
    entryPoints: ["src/renderer/index.tsx"],
    bundle: true,
    outdir: "dist",
    entryNames: "renderer",
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
