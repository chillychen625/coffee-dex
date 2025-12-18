import { contextBridge, ipcRenderer } from "electron";

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld("electron", {
  getAppVersion: () => ipcRenderer.invoke("get-app-version"),
  showError: (title: string, message: string) =>
    ipcRenderer.invoke("show-error", title, message),
});

// Type definitions for the exposed API
declare global {
  interface Window {
    electron: {
      getAppVersion: () => Promise<string>;
      showError: (title: string, message: string) => Promise<void>;
    };
  }
}

export {};
