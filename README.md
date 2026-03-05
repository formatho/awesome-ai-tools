# Agent Orchestrator

**Desktop application for managing autonomous AI agents locally.**

---

## 🚀 Quick Start

### Option 1: Using Startup Script

```bash
# Start both backend + frontend
./start.sh

# Or start individually
./start.sh backend   # API server only
./start.sh frontend  # Electron app only
```

### Option 2: Manual Start

**Terminal 1 - Backend:**
```bash
cd backend
go build -o bin/server ./cmd/server
./bin/server
```

**Terminal 2 - Frontend:**
```bash
cd electron-app
npm install  # First time only
npm run dev
```

---

## 📦 What You Get

- **Backend API** running on `http://localhost:18765`
- **Electron App** with React UI
- **SQLite Database** for persistence
- **WebSocket** for real-time updates

---

## 🎯 Features

### 6 Modular Go Libraries
1. **go-llm-client** - Unified LLM interface (OpenAI, Anthropic, Ollama)
2. **go-agent-pool** - Agent lifecycle management
3. **go-agent-skills** - Skill execution with permissions
4. **go-todo-queue** - Persistent TODO management
5. **go-cron-agents** - Cron scheduling for agents
6. **go-agent-config** - Configuration management

### Backend API
- 44 REST endpoints
- WebSocket support
- SQLite database
- All libraries integrated

### Electron UI
- Dark theme
- Real-time updates
- Dashboard, Agents, TODOs, Cron, Config views

---

## 🛠️ Tech Stack

**Backend:**
- Go 1.24
- Fiber (web framework)
- Gorilla WebSocket
- SQLite

**Frontend:**
- Electron
- React 18
- TypeScript
- Tailwind CSS
- React Query

---

## 📊 Project Stats

```
Libraries:     6 (24,000 lines)
Backend:       19 files (2,500 lines)
Frontend:      25+ files (2,500 lines)
Tests:         200+ passing
Total Lines:   ~29,000
```

---

## 🎨 Screenshots

Coming soon...

---

## 📝 API Endpoints

### Agents
- `GET/POST /api/agents` - List/Create agents
- `GET/PUT/DELETE /api/agents/:id` - Agent CRUD
- `POST /api/agents/:id/pause` - Pause agent
- `POST /api/agents/:id/resume` - Resume agent

### TODOs
- `GET/POST /api/todos` - List/Create TODOs
- `GET/PUT/DELETE /api/todos/:id` - TODO CRUD
- `POST /api/todos/:id/start` - Start TODO
- `POST /api/todos/:id/cancel` - Cancel TODO

### Cron
- `GET/POST /api/cron` - List/Create cron jobs
- `GET/PUT/DELETE /api/cron/:id` - Cron CRUD
- `POST /api/cron/:id/pause` - Pause cron
- `POST /api/cron/:id/resume` - Resume cron
- `GET /api/cron/:id/history` - Run history

### Config
- `GET/PUT /api/config` - Configuration management
- `POST /api/config/test-llm` - Test LLM connection

### System
- `GET /health` - Health check
- `GET /api/system/status` - System status

### WebSocket
- `GET /ws` - WebSocket endpoint for real-time updates

---

## 🔧 Development

### Prerequisites
- Go 1.24+
- Node.js 18+
- npm or yarn

### Build Backend
```bash
cd backend
go mod tidy
go build -o bin/server ./cmd/server
```

### Build Frontend
```bash
cd electron-app
npm install
npm run build
```

### Run Tests
```bash
# Backend
cd backend
go test ./...

# Libraries
cd packages/go-llm-client
go test ./...
```

---

## 📄 License

MIT License - See LICENSE file for details

---

## 🤝 Contributing

Contributions welcome! Please read CONTRIBUTING.md first.

---

## 📧 Contact

- **GitHub:** https://github.com/formatho/agent-orchestrator
- **Website:** https://formatho.com

---

**Built with ❤️ by Formatho**
