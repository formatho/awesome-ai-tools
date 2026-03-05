import { app, BrowserWindow, ipcMain } from 'electron'
import path from 'path'
import { fileURLToPath } from 'url'
import { setupTray } from './tray'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// The built directory structure
//
// ├─ dist-electron/
// │  ├─ main.js
// │  └─ preload.js
// └─ dist/
//    └─ index.html

process.env.DIST_ELECTRON = path.join(__dirname, '../')
process.env.DIST = path.join(__dirname, '../dist')
process.env.VITE_PUBLIC = process.env.VITE_PUBLIC || path.join(__dirname, '../public')

let win: BrowserWindow | null = null

// Here, you can also use other preload
const preload = path.join(__dirname, './preload.js')

function createWindow() {
  win = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1000,
    minHeight: 700,
    title: 'Agent Orchestrator',
    icon: path.join(process.env.VITE_PUBLIC || '', 'icon.png'),
    backgroundColor: '#0f0f0f',
    frame: false,
    titleBarStyle: 'hiddenInset',
    webPreferences: {
      preload,
      nodeIntegration: false,
      contextIsolation: true,
    },
  })

  // Test active push message to Renderer-process
  win.webContents.on('did-finish-load', () => {
    win?.webContents.send('main-process-message', new Date().toLocaleString())
  })

  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL)
    win.webContents.openDevTools()
  } else {
    win.loadFile(path.join(process.env.DIST || '', 'index.html'))
  }

  return win
}

// Quit when all windows are closed, except on macOS
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
    win = null
  }
})

app.on('activate', () => {
  // On macOS, re-create window when dock icon is clicked
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow()
  }
})

// Handle window controls from renderer
ipcMain.on('window-minimize', () => {
  win?.minimize()
})

ipcMain.on('window-maximize', () => {
  if (win?.isMaximized()) {
    win.unmaximize()
  } else {
    win?.maximize()
  }
})

ipcMain.on('window-close', () => {
  win?.close()
})

ipcMain.handle('window-is-maximized', () => {
  return win?.isMaximized() || false
})

app.whenReady().then(() => {
  createWindow()
  setupTray(win)
})
