import { contextBridge, ipcRenderer } from "electron";

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld("electron", {
  getAppVersion: () => ipcRenderer.invoke("get-app-version"),
  showError: (title: string, message: string) =>
    ipcRenderer.invoke("show-error", title, message),
  minimizeWindow: () => ipcRenderer.send("minimize-window"),
  closeWindow: () => ipcRenderer.send("close-window"),
  getAnthropicApiKey: () => ipcRenderer.invoke("get-anthropic-api-key"),
  toggleAlwaysOnTop: () => ipcRenderer.send("toggle-always-on-top"),
  getAlwaysOnTop: () => ipcRenderer.invoke("get-always-on-top"),
  onAlwaysOnTopChanged: (callback: (value: boolean) => void) => {
    ipcRenderer.on("always-on-top-changed", (_event, value) => callback(value));
  },
});

// Type definitions for the exposed API
declare global {
  interface Window {
    electron: {
      getAppVersion: () => Promise<string>;
      showError: (title: string, message: string) => Promise<void>;
      minimizeWindow: () => void;
      closeWindow: () => void;
      getAnthropicApiKey: () => Promise<string | null>;
      toggleAlwaysOnTop: () => void;
      getAlwaysOnTop: () => Promise<boolean>;
      onAlwaysOnTopChanged: (callback: (value: boolean) => void) => void;
    };
  }
}

export {};
