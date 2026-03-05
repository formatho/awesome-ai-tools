import { Tray, Menu, nativeImage, BrowserWindow, app } from 'electron'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

let tray: Tray | null = null

export function setupTray(win: BrowserWindow | null) {
  // Create tray icon
  const iconPath = path.join(__dirname, '../public/icon.png')
  const icon = nativeImage.createFromPath(iconPath)
  
  // Use a simple 16x16 template icon for macOS
  const trayIcon = icon.resize({ width: 16, height: 16 })
  
  tray = new Tray(trayIcon)
  
  const contextMenu = Menu.buildFromTemplate([
    {
      label: 'Open Agent Orchestrator',
      click: () => {
        if (win) {
          win.show()
          win.focus()
        }
      }
    },
    {
      label: 'Dashboard',
      click: () => {
        if (win) {
          win.show()
          win.webContents.send('navigate', '/')
        }
      }
    },
    { type: 'separator' },
    {
      label: 'Agents',
      click: () => {
        if (win) {
          win.show()
          win.webContents.send('navigate', '/agents')
        }
      }
    },
    {
      label: 'TODOs',
      click: () => {
        if (win) {
          win.show()
          win.webContents.send('navigate', '/todos')
        }
      }
    },
    { type: 'separator' },
    {
      label: 'Quit',
      click: () => {
        app.quit()
      }
    }
  ])
  
  tray.setToolTip('Agent Orchestrator')
  tray.setContextMenu(contextMenu)
  
  // Show window on click
  tray.on('click', () => {
    if (win) {
      if (win.isVisible()) {
        win.hide()
      } else {
        win.show()
        win.focus()
      }
    }
  })
  
  return tray
}
