// Package api provides the HTTP server and routing.
package api

import (
	"database/sql"

	fastws "github.com/fasthttp/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/api/handlers"
	"github.com/formatho/agent-orchestrator/backend/internal/api/websocket"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/valyala/fasthttp"
)

// Server represents the API server.
type Server struct {
	app *fiber.App
	db  *sql.DB
	hub *websocket.Hub

	// Services
	agentSvc  *services.AgentService
	todoSvc   *services.TODOService
	cronSvc   *services.CronService
	configSvc *services.ConfigService
	chatSvc   *services.ChatService

	// Handlers
	agentH  *handlers.AgentHandler
	todoH   *handlers.TODOHandler
	cronH   *handlers.CronHandler
	configH *handlers.ConfigHandler
	systemH *handlers.SystemHandler
	chatH   *handlers.ChatHandler

	// WebSocket upgrader
	upgrader fastws.FastHTTPUpgrader
}

// NewServer creates a new API server.
func NewServer(db *sql.DB) *fiber.App {
	// Create WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create services
	agentSvc := services.NewAgentService(db, hub)
	todoSvc := services.NewTODOService(db, hub)
	cronSvc := services.NewCronService(db, hub)
	configSvc := services.NewConfigService(db)
	chatSvc := services.NewChatService(db, agentSvc, configSvc)

	// Create handlers
	agentH := handlers.NewAgentHandler(agentSvc)
	todoH := handlers.NewTODOHandler(todoSvc)
	cronH := handlers.NewCronHandler(cronSvc)
	configH := handlers.NewConfigHandler(configSvc)
	systemH := handlers.NewSystemHandler(agentSvc, todoSvc, cronSvc)
	chatH := handlers.NewChatHandler(chatSvc)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Agent Orchestrator API v1.0.0",
		ServerHeader: "Agent-Orchestrator",
	})

	// Create server instance for WebSocket
	server := &Server{
		app:       app,
		db:        db,
		hub:       hub,
		agentSvc:  agentSvc,
		todoSvc:   todoSvc,
		cronSvc:   cronSvc,
		configSvc: configSvc,
		chatSvc:   chatSvc,
		agentH:    agentH,
		todoH:     todoH,
		cronH:     cronH,
		configH:   configH,
		systemH:   systemH,
		chatH:     chatH,
		upgrader: fastws.FastHTTPUpgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
				return true // Allow all origins for development
			},
		},
	}

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup routes
	server.setupRoutes()

	return app
}

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	// Health check
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": c.Context().ConnTime(),
		})
	})

	// WebSocket endpoint
	s.app.Get("/ws", func(c *fiber.Ctx) error {
		return s.upgrader.Upgrade(c.Context(), func(conn *fastws.Conn) {
			// Create client and register with hub
			client := websocket.NewClient(s.hub, conn)
			s.hub.Register(client)

			// Start read/write pumps
			go client.WritePump()
			client.ReadPump()
		})
	})

	// API routes
	api := s.app.Group("/api")

	// Agent routes
	agents := api.Group("/agents")
	agents.Get("/", s.agentH.List)
	agents.Post("/", s.agentH.Create)
	agents.Get("/:id", s.agentH.Get)
	agents.Put("/:id", s.agentH.Update)
	agents.Delete("/:id", s.agentH.Delete)
	agents.Post("/:id/start", s.agentH.Start)
	agents.Post("/:id/stop", s.agentH.Stop)
	agents.Post("/:id/pause", s.agentH.Pause)
	agents.Post("/:id/resume", s.agentH.Resume)
	// Chat routes (nested under agent)
	agents.Get("/:id/chat", s.chatH.GetHistory)
	agents.Post("/:id/chat", s.chatH.SendMessage)
	agents.Delete("/:id/chat", s.chatH.ClearHistory)

	// TODO routes
	todos := api.Group("/todos")
	todos.Get("/", s.todoH.List)
	todos.Post("/", s.todoH.Create)
	todos.Get("/:id", s.todoH.Get)
	todos.Put("/:id", s.todoH.Update)
	todos.Delete("/:id", s.todoH.Delete)
	todos.Post("/:id/start", s.todoH.Start)
	todos.Post("/:id/cancel", s.todoH.Cancel)

	// Cron routes
	cron := api.Group("/cron")
	cron.Get("/", s.cronH.List)
	cron.Post("/", s.cronH.Create)
	cron.Get("/:id", s.cronH.Get)
	cron.Put("/:id", s.cronH.Update)
	cron.Delete("/:id", s.cronH.Delete)
	cron.Post("/:id/pause", s.cronH.Pause)
	cron.Post("/:id/resume", s.cronH.Resume)
	cron.Get("/:id/history", s.cronH.GetHistory)

	// Config routes
	config := api.Group("/config")
	config.Get("/", s.configH.Get)
	config.Put("/", s.configH.Update)
	config.Post("/test-llm", s.configH.TestLLM)

	// System routes
	system := api.Group("/system")
	system.Get("/status", s.systemH.Status)
	system.Get("/health", s.systemH.Health)
}
