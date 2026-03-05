import { app, BrowserWindow, ipcMain } from 'electron'
import path from 'path'
import { fileURLToPath } from 'url'
import { spawn, ChildProcess } from 'child_process'
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
let backendProcess: ChildProcess | null = null

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
    stopBackend()
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

// Start the Go backend process
function startBackend(): Promise<void> {
  return new Promise((resolve, reject) => {
    // Skip backend startup if BACKEND_URL is set (external backend running)
    if (process.env.BACKEND_URL) {
      console.log('🔗 Using external backend:', process.env.BACKEND_URL)
      return resolve()
    }
    
    // In dev mode without BACKEND_URL, assume backend is managed separately
    const isDev = !app.isPackaged
    if (isDev && !process.env.START_BACKEND) {
      console.log('🔧 Dev mode: assuming backend is running separately')
      return resolve()
    }
    
    // In production: backend is at process.resourcesPath/backend/server
    const backendExe = isDev
      ? path.join(__dirname, '../../backend/bin/server')
      : path.join(process.resourcesPath, 'backend', 'server')
    
    const userDataPath = app.getPath('userData')
    const dbPath = path.join(userDataPath, 'agent-orchestrator.db')
    const port = ':18765'
    
    console.log('🚀 Starting backend...')
    console.log('   Binary:', backendExe)
    console.log('   DB Path:', dbPath)
    console.log('   Port:', port)
    
    backendProcess = spawn(backendExe, [], {
      env: {
        ...process.env,
        DB_PATH: dbPath,
        PORT: port,
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    })
    
    backendProcess.stdout?.on('data', (data) => {
      console.log(`[Backend] ${data.toString().trim()}`)
    })
    
    backendProcess.stderr?.on('data', (data) => {
      console.error(`[Backend Error] ${data.toString().trim()}`)
    })
    
    backendProcess.on('error', (err) => {
      console.error('Failed to start backend:', err)
      reject(err)
    })
    
    // Give backend a moment to start
    setTimeout(() => {
      if (backendProcess && !backendProcess.killed) {
        console.log('✅ Backend started successfully')
        resolve()
      } else {
        reject(new Error('Backend process died immediately'))
      }
    }, 1000)
  })
}

// Stop the backend process
function stopBackend() {
  if (backendProcess) {
    console.log('🛑 Stopping backend...')
    backendProcess.kill()
    backendProcess = null
  }
}

app.whenReady().then(async () => {
  try {
    await startBackend()
    createWindow()
    setupTray(win)
  } catch (err) {
    console.error('Failed to start app:', err)
    app.quit()
  }
})

// Cleanup on app quit
app.on('will-quit', () => {
  stopBackend()
})
