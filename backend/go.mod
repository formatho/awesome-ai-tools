module github.com/formatho/agent-orchestrator/backend

go 1.24

require (
	github.com/fasthttp/websocket v1.5.12
	github.com/formatho/agent-orchestrator/packages/agent-config v0.0.0
	github.com/formatho/agent-orchestrator/packages/agent-pool v0.0.0
	github.com/formatho/agent-orchestrator/packages/cron-agents v0.0.0
	github.com/formatho/agent-orchestrator/packages/llm-client v0.0.0
	github.com/formatho/agent-orchestrator/packages/todo-queue v0.0.0
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.22
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.58.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/formatho/agent-orchestrator/packages/agent-config => ../packages/agent-config
	github.com/formatho/agent-orchestrator/packages/agent-pool => ../packages/agent-pool
	github.com/formatho/agent-orchestrator/packages/agent-skills => ../packages/agent-skills
	github.com/formatho/agent-orchestrator/packages/cron-agents => ../packages/cron-agents
	github.com/formatho/agent-orchestrator/packages/llm-client => ../packages/llm-client
	github.com/formatho/agent-orchestrator/packages/todo-queue => ../packages/todo-queue
)
