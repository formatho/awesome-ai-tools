import { nativeImage as w, Tray as T, Menu as v, app as s, BrowserWindow as f, ipcMain as d } from "electron";
import n from "path";
import { fileURLToPath as h } from "url";
import { spawn as E } from "child_process";
const D = n.dirname(h(import.meta.url));
let r = null;
function P(e) {
  const c = n.join(D, "../public/icon.png"), l = w.createFromPath(c).resize({ width: 16, height: 16 });
  r = new T(l);
  const m = v.buildFromTemplate([
    {
      label: "Open Agent Orchestrator",
      click: () => {
        e && (e.show(), e.focus());
      }
    },
    {
      label: "Dashboard",
      click: () => {
        e && (e.show(), e.webContents.send("navigate", "/"));
      }
    },
    { type: "separator" },
    {
      label: "Agents",
      click: () => {
        e && (e.show(), e.webContents.send("navigate", "/agents"));
      }
    },
    {
      label: "TODOs",
      click: () => {
        e && (e.show(), e.webContents.send("navigate", "/todos"));
      }
    },
    { type: "separator" },
    {
      label: "Quit",
      click: () => {
        s.quit();
      }
    }
  ]);
  return r.setToolTip("Agent Orchestrator"), r.setContextMenu(m), r.on("click", () => {
    e && (e.isVisible() ? e.hide() : (e.show(), e.focus()));
  }), r;
}
const a = n.dirname(h(import.meta.url));
process.env.DIST_ELECTRON = n.join(a, "../");
process.env.DIST = n.join(a, "../dist");
process.env.VITE_PUBLIC = process.env.VITE_PUBLIC || n.join(a, "../public");
let o = null, t = null;
const _ = n.join(a, "./preload.js");
function b() {
  return o = new f({
    width: 1400,
    height: 900,
    minWidth: 1e3,
    minHeight: 700,
    title: "Agent Orchestrator",
    icon: n.join(process.env.VITE_PUBLIC || "", "icon.png"),
    backgroundColor: "#0f0f0f",
    frame: !1,
    titleBarStyle: "hiddenInset",
    webPreferences: {
      preload: _,
      nodeIntegration: !1,
      contextIsolation: !0
    }
  }), o.webContents.on("did-finish-load", () => {
    o?.webContents.send("main-process-message", (/* @__PURE__ */ new Date()).toLocaleString());
  }), process.env.VITE_DEV_SERVER_URL ? (o.loadURL(process.env.VITE_DEV_SERVER_URL), o.webContents.openDevTools()) : o.loadFile(n.join(process.env.DIST || "", "index.html")), o;
}
s.on("window-all-closed", () => {
  process.platform !== "darwin" && (k(), s.quit(), o = null);
});
s.on("activate", () => {
  f.getAllWindows().length === 0 && b();
});
d.on("window-minimize", () => {
  o?.minimize();
});
d.on("window-maximize", () => {
  o?.isMaximized() ? o.unmaximize() : o?.maximize();
});
d.on("window-close", () => {
  o?.close();
});
d.handle("window-is-maximized", () => o?.isMaximized() || !1);
function B() {
  return new Promise((e, c) => {
    if (process.env.BACKEND_URL)
      return console.log("🔗 Using external backend:", process.env.BACKEND_URL), e();
    const p = !s.isPackaged;
    if (p && !process.env.START_BACKEND)
      return console.log("🔧 Dev mode: assuming backend is running separately"), e();
    const l = p ? n.join(a, "../../backend/bin/server") : n.join(process.resourcesPath, "backend", "server"), m = s.getPath("userData"), u = n.join(m, "agent-orchestrator.db"), g = ":18765";
    console.log("🚀 Starting backend..."), console.log("   Binary:", l), console.log("   DB Path:", u), console.log("   Port:", g), t = E(l, [], {
      env: {
        ...process.env,
        DB_PATH: u,
        PORT: g
      },
      stdio: ["ignore", "pipe", "pipe"]
    }), t.stdout?.on("data", (i) => {
      console.log(`[Backend] ${i.toString().trim()}`);
    }), t.stderr?.on("data", (i) => {
      console.error(`[Backend Error] ${i.toString().trim()}`);
    }), t.on("error", (i) => {
      console.error("Failed to start backend:", i), c(i);
    }), setTimeout(() => {
      t && !t.killed ? (console.log("✅ Backend started successfully"), e()) : c(new Error("Backend process died immediately"));
    }, 1e3);
  });
}
function k() {
  t && (console.log("🛑 Stopping backend..."), t.kill(), t = null);
}
s.whenReady().then(async () => {
  try {
    await B(), b(), P(o);
  } catch (e) {
    console.error("Failed to start app:", e), s.quit();
  }
});
s.on("will-quit", () => {
  k();
});
