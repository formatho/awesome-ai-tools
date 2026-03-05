import { nativeImage, Tray, Menu, app, BrowserWindow, ipcMain } from "electron";
import path from "path";
import { fileURLToPath } from "url";
const __dirname$2 = path.dirname(fileURLToPath(import.meta.url));
let tray = null;
function setupTray(win2) {
  const iconPath = path.join(__dirname$2, "../public/icon.png");
  const icon = nativeImage.createFromPath(iconPath);
  const trayIcon = icon.resize({ width: 16, height: 16 });
  tray = new Tray(trayIcon);
  const contextMenu = Menu.buildFromTemplate([
    {
      label: "Open Agent Orchestrator",
      click: () => {
        if (win2) {
          win2.show();
          win2.focus();
        }
      }
    },
    {
      label: "Dashboard",
      click: () => {
        if (win2) {
          win2.show();
          win2.webContents.send("navigate", "/");
        }
      }
    },
    { type: "separator" },
    {
      label: "Agents",
      click: () => {
        if (win2) {
          win2.show();
          win2.webContents.send("navigate", "/agents");
        }
      }
    },
    {
      label: "TODOs",
      click: () => {
        if (win2) {
          win2.show();
          win2.webContents.send("navigate", "/todos");
        }
      }
    },
    { type: "separator" },
    {
      label: "Quit",
      click: () => {
        app.quit();
      }
    }
  ]);
  tray.setToolTip("Agent Orchestrator");
  tray.setContextMenu(contextMenu);
  tray.on("click", () => {
    if (win2) {
      if (win2.isVisible()) {
        win2.hide();
      } else {
        win2.show();
        win2.focus();
      }
    }
  });
  return tray;
}
const __dirname$1 = path.dirname(fileURLToPath(import.meta.url));
process.env.DIST_ELECTRON = path.join(__dirname$1, "../");
process.env.DIST = path.join(__dirname$1, "../dist");
process.env.VITE_PUBLIC = process.env.VITE_PUBLIC || path.join(__dirname$1, "../public");
let win = null;
const preload = path.join(__dirname$1, "./preload.js");
function createWindow() {
  win = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1e3,
    minHeight: 700,
    title: "Agent Orchestrator",
    icon: path.join(process.env.VITE_PUBLIC || "", "icon.png"),
    backgroundColor: "#0f0f0f",
    frame: false,
    titleBarStyle: "hiddenInset",
    webPreferences: {
      preload,
      nodeIntegration: false,
      contextIsolation: true
    }
  });
  win.webContents.on("did-finish-load", () => {
    win?.webContents.send("main-process-message", (/* @__PURE__ */ new Date()).toLocaleString());
  });
  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL);
    win.webContents.openDevTools();
  } else {
    win.loadFile(path.join(process.env.DIST || "", "index.html"));
  }
  return win;
}
app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
    win = null;
  }
});
app.on("activate", () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});
ipcMain.on("window-minimize", () => {
  win?.minimize();
});
ipcMain.on("window-maximize", () => {
  if (win?.isMaximized()) {
    win.unmaximize();
  } else {
    win?.maximize();
  }
});
ipcMain.on("window-close", () => {
  win?.close();
});
ipcMain.handle("window-is-maximized", () => {
  return win?.isMaximized() || false;
});
app.whenReady().then(() => {
  createWindow();
  setupTray(win);
});
