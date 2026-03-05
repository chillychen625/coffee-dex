import { app, BrowserWindow, Menu, ipcMain } from "electron";
import * as path from "path";
import { spawn, ChildProcess } from "child_process";
import * as fs from "fs";
import * as dotenv from "dotenv";

// Load .env file from project root (two levels up from dist/)
const envPath = path.join(__dirname, "../..", ".env");
if (fs.existsSync(envPath)) {
  dotenv.config({ path: envPath });
  console.log("[Config] Loaded .env file");
}

let mainWindow: BrowserWindow | null = null;
let backendProcess: ChildProcess | null = null;

function startBackend(): void {
  // __dirname is dist/, go up two levels to reach the Go project root
  const goProjectDir = path.join(__dirname, "../..");

  // Check for pre-built binary first (for production), then fall back to go run
  const binaryPath = path.join(goProjectDir, "bin", "coffee-dex");
  const useBinary = fs.existsSync(binaryPath);

  console.log(`[Backend] Starting Go backend from: ${goProjectDir}`);

  if (useBinary) {
    console.log(`[Backend] Using pre-built binary: ${binaryPath}`);
    backendProcess = spawn(binaryPath, ["-storage=mysql"], {
      cwd: goProjectDir,
      stdio: ["ignore", "pipe", "pipe"],
    });
  } else {
    console.log(`[Backend] Using 'go run main.go'`);
    backendProcess = spawn("go", ["run", "main.go", "-storage=mysql"], {
      cwd: goProjectDir,
      stdio: ["ignore", "pipe", "pipe"],
      env: { ...process.env },
    });
  }

  backendProcess.stdout?.on("data", (data: Buffer) => {
    console.log(`[Backend] ${data.toString().trim()}`);
  });

  backendProcess.stderr?.on("data", (data: Buffer) => {
    console.error(`[Backend Error] ${data.toString().trim()}`);
  });

  backendProcess.on("error", (err: Error) => {
    console.error(`[Backend] Failed to start: ${err.message}`);
  });

  backendProcess.on("exit", (code: number | null, signal: string | null) => {
    console.log(`[Backend] Process exited with code ${code}, signal ${signal}`);
    backendProcess = null;
  });
}

function stopBackend(): void {
  if (backendProcess) {
    console.log("[Backend] Stopping backend process...");
    backendProcess.kill("SIGTERM");
    backendProcess = null;
  }
}

function createWindow(): void {
  const { screen } = require("electron");
  const primaryDisplay = screen.getPrimaryDisplay();
  const workArea = primaryDisplay.workAreaSize;

  mainWindow = new BrowserWindow({
    width: 640,
    height: workArea.height,
    minWidth: 640,
    maxWidth: 640,
    minHeight: 400,
    x: workArea.width - 640, // Pin to right edge
    y: 0,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, "preload.js"),
    },
    frame: false, // Custom frame with title bar
    transparent: false,
    backgroundColor: "#9bbc0f",
    title: "CoffeeDex",
    icon: path.join(__dirname, "static/icon.png"),
    resizable: true,
  });

  mainWindow.loadFile("dist/index.html");

  mainWindow.on("closed", () => {
    mainWindow = null;
  });

  // Create application menu
  createMenu();
}

function createMenu(): void {
  const template: Electron.MenuItemConstructorOptions[] = [
    {
      label: "CoffeeDex",
      submenu: [
        {
          label: "About CoffeeDex",
          click: () => {
            const { dialog } = require("electron");
            dialog.showMessageBox(mainWindow!, {
              type: "info",
              title: "About CoffeeDex",
              message: "CoffeeDex v1.0.0",
              detail:
                "A Pokemon-themed coffee logging desktop application.\nTransform your coffee tasting notes into Pokemon!",
            });
          },
        },
        { type: "separator" },
        {
          label: "Quit",
          accelerator: "CmdOrCtrl+Q",
          click: () => {
            app.quit();
          },
        },
      ],
    },
    {
      label: "View",
      submenu: [
        {
          label: "Toggle Fullscreen",
          accelerator: "F11",
          click: () => {
            if (mainWindow) {
              mainWindow.setFullScreen(!mainWindow.isFullScreen());
            }
          },
        },
        {
          label: "Reload",
          accelerator: "F5",
          click: () => {
            if (mainWindow) {
              mainWindow.reload();
            }
          },
        },
        { type: "separator" },
        {
          label: "Developer Tools",
          accelerator: "F12",
          click: () => {
            if (mainWindow) {
              mainWindow.webContents.toggleDevTools();
            }
          },
        },
      ],
    },
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}

app.whenReady().then(() => {
  startBackend();
  createWindow();
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

app.on("before-quit", () => {
  stopBackend();
});

app.on("will-quit", () => {
  stopBackend();
});

app.on("activate", () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

// IPC handlers
ipcMain.handle("get-app-version", () => {
  return app.getVersion();
});

ipcMain.handle("show-error", (event: any, title: string, message: string) => {
  const { dialog } = require("electron");
  return dialog.showErrorBox(title, message);
});

// Window control IPC handlers
ipcMain.on("minimize-window", () => {
  if (mainWindow) {
    mainWindow.minimize();
  }
});

ipcMain.on("close-window", () => {
  if (mainWindow) {
    mainWindow.close();
  }
});

ipcMain.on("toggle-always-on-top", () => {
  if (mainWindow) {
    const current = mainWindow.isAlwaysOnTop();
    mainWindow.setAlwaysOnTop(!current);
    mainWindow.webContents.send("always-on-top-changed", !current);
  }
});

ipcMain.handle("get-always-on-top", () => {
  return mainWindow?.isAlwaysOnTop() ?? false;
});

// Get API key from environment
ipcMain.handle("get-anthropic-api-key", () => {
  return process.env.ANTHROPIC_API_KEY || null;
});
