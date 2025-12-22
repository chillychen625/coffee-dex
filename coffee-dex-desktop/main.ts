import { app, BrowserWindow, Menu, ipcMain } from "electron";
import * as path from "path";

let mainWindow: BrowserWindow | null = null;

function createWindow(): void {
  mainWindow = new BrowserWindow({
    width: 640,
    height: 640,
    minWidth: 640,
    minHeight: 640,
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

app.whenReady().then(createWindow);

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
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
