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
	agentSvc       *services.AgentService
	todoSvc        *services.TODOService
	cronSvc        *services.CronService
	configSvc      *services.ConfigService
	chatSvc        *services.ChatService
	orgSvc         *services.OrgService
	stateSvc       *services.StateService
	newsletterSvc  *services.NewsletterService
	betaSvc        *services.BetaSignupService
	betaFeedbackSvc *services.BetaFeedbackService

	// Handlers
	agentH       *handlers.AgentHandler
	todoH        *handlers.TODOHandler
	cronH        *handlers.CronHandler
	configH      *handlers.ConfigHandler
	systemH      *handlers.SystemHandler
	chatH        *handlers.ChatHandler
	orgH         *handlers.OrgHandler
	stateH       *handlers.StateHandler
	authH        *handlers.AuthHandlerFiber
	teamInvH     *handlers.TeamInvitationHandler
	teamPermH    *handlers.TeamPermissionsHandler
	stripeH      *handlers.StripeHandler
	analyticsH   *handlers.AnalyticsHandler
	newsletterH  *handlers.NewsletterHandler
	betaH        *handlers.BetaSignupHandler
	betaFeedbackH *handlers.BetaFeedbackHandler

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
	orgSvc := services.NewOrgService(db)
	stateSvc := services.NewStateService(db)

	// Create handlers
	agentH := handlers.NewAgentHandler(agentSvc)
	todoH := handlers.NewTODOHandler(todoSvc)
	cronH := handlers.NewCronHandler(cronSvc)
	configH := handlers.NewConfigHandler(configSvc)
	systemH := handlers.NewSystemHandler(agentSvc, todoSvc, cronSvc)
	chatH := handlers.NewChatHandler(chatSvc)
	orgH := handlers.NewOrgHandler(orgSvc)
	stateH := handlers.NewStateHandler(stateSvc)
	authH := handlers.NewAuthHandlerFiber()

	// Create auth service for team handlers and Fiber auth handler
	authSvc := services.NewAuthService(db)

	// Create subscription service
	subSvc := services.NewSubscriptionService(db)

	// Create analytics service
	analyticsSvc := services.NewAnalyticsService(db)

	// Create newsletter service
	newsletterSvc := services.NewNewsletterService(db)

	// Create beta signup service
	betaSvc := services.NewBetaSignupService(db)

	// Create beta feedback service
	betaFeedbackSvc := services.NewBetaFeedbackService(db)

	// Create team handlers
	mockEmailSender := &services.MockEmailSender{}
	invitationSvc := services.NewInvitationService(db, mockEmailSender)
	permissionSvc := services.NewPermissionService(db)

	// Create Stripe handler with subscription service and analytics
	stripeH := handlers.NewStripeHandler(authSvc, subSvc, analyticsSvc)

	// Create analytics handler
	analyticsH := handlers.NewAnalyticsHandler(analyticsSvc, subSvc)

	// Create newsletter handler
	newsletterH := handlers.NewNewsletterHandler(newsletterSvc)

	// Create beta signup handler
	betaH := handlers.NewBetaSignupHandler(betaSvc)

	// Create beta feedback handler
	betaFeedbackH := handlers.NewBetaFeedbackHandler(betaFeedbackSvc)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Agent Orchestrator API v1.0.0",
		ServerHeader: "Agent-Orchestrator",
	})

	// Create server instance
	server := &Server{
		app:           app,
		db:            db,
		hub:           hub,
		agentSvc:      agentSvc,
		todoSvc:       todoSvc,
		cronSvc:       cronSvc,
		configSvc:     configSvc,
		chatSvc:       chatSvc,
		orgSvc:        orgSvc,
		newsletterSvc: newsletterSvc,
		betaSvc:       betaSvc,
		betaFeedbackSvc: betaFeedbackSvc,
		stateH:        stateH,
		authH:         authH,
		agentH:        agentH,
		todoH:         todoH,
		cronH:         cronH,
		configH:       configH,
		systemH:       systemH,
		chatH:         chatH,
		orgH:          orgH,
		teamInvH:      handlers.NewTeamInvitationHandler(invitationSvc, authSvc),
		teamPermH:     handlers.NewTeamPermissionsHandler(permissionSvc, authSvc),
		stripeH:       stripeH,
		analyticsH:    analyticsH,
		newsletterH:   newsletterH,
		betaH:         betaH,
		betaFeedbackH: betaFeedbackH,
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
	agents.Get("/:id/logs", s.agentH.Logs)
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

	// Auth endpoints - Login is public, others require authentication
	auth := api.Group("/auth")
	
	// Public login endpoint (no auth required)
	auth.Post("/login", s.authH.Login)
	
	// Protected auth routes (require authentication for logout/refresh/me)
	protectedAuth := api.Group("/auth").Use(s.authH.RequireAuth())
	protectedAuth.Post("/logout", s.authH.Logout)
	protectedAuth.Post("/refresh", s.authH.RefreshToken)
	protectedAuth.Get("/me", s.authH.CurrentUser)

	// Organization routes
	orgs := api.Group("/organizations")
	orgs.Get("/", s.orgH.List)
	orgs.Post("/", s.orgH.Create)
	orgs.Get("/:id", s.orgH.Get)
	orgs.Put("/:id", s.orgH.Update)
	orgs.Delete("/:id", s.orgH.Delete)
	orgs.Get("/slug/:slug", s.orgH.GetBySlug)
	orgs.Get("/owner/:ownerId", s.orgH.GetByOwner)
	orgs.Post("/switch", s.orgH.SwitchOrganization)

	// Organization member management
	members := orgs.Group("/:id/members")
	members.Get("/", s.orgH.ListMembers)
	members.Post("/", s.orgH.InviteMember)
	members.Get("/:userId", s.orgH.GetMember)
	members.Patch("/:userId/role", s.orgH.UpdateMemberRole)
	members.Delete("/:userId", s.orgH.RemoveMember)

	// Kick endpoint (alias for remove with explicit action)
	orgs.Delete("/:id/members/:userId/kick", s.orgH.KickMember)

	// Team invitations (manual registration)
	teamInv := api.Group("/team/invitations")
	teamInv.Post("/", s.teamInvH.CreateInvitation)
	teamInv.Get("/", s.teamInvH.ListInvitations)
	teamInv.Get("/:id", s.teamInvH.GetInvitation)
	teamInv.Delete("/:id", s.teamInvH.CancelInvitation)
	teamInv.Post("/accept", s.teamInvH.AcceptInvitation)
	teamInv.Post("/reject/:id", s.teamInvH.RejectInvitation)
	teamInv.Get("/verify-token", s.teamInvH.VerifyToken)
	teamInv.Get("/stats", s.teamInvH.GetStats)

	// Team permissions (manual registration)
	teamPerm := api.Group("/team/permissions")
	teamPerm.Get("/check", s.teamPermH.CheckPermission)
	teamPerm.Post("/grant", s.teamPermH.GrantPermission)
	teamPerm.Delete("/revoke", s.teamPermH.RevokePermission)
	teamPerm.Get("/user/:userId/org/:orgId", s.teamPermH.GetUserPermissions)
	teamPerm.Get("/resource/:orgId/resource/:resource", s.teamPermH.GetResourcePermissions)
	teamPerm.Get("/bulk-check", s.teamPermH.BulkCheckPermissions)
	teamPerm.Post("/templates", s.teamPermH.CreatePermissionTemplate)
	teamPerm.Get("/templates", s.teamPermH.ListPermissionTemplates)
	teamPerm.Get("/templates/:id", s.teamPermH.GetPermissionTemplate)
	teamPerm.Put("/templates/:id", s.teamPermH.UpdatePermissionTemplate)
	teamPerm.Delete("/templates/:id", s.teamPermH.DeletePermissionTemplate)

	// State Persistence Management (NEW - Phase 4 Feature)
	state := api.Group("/agent-state")
	state.Post("/", s.stateH.SaveState)
	state.Get("/:agentID", s.stateH.GetAgentState)
	state.Get("/:agentID/history", s.stateH.GetAgentStateHistory)
	state.Patch("/:agentID", s.stateH.UpdateAgentState)
	state.Delete("/:agentID", s.stateH.DeleteAgentState)
	state.Get("/:agentID/export", s.stateH.ExportAgentState)

	// State summary and metrics endpoints
	api.Get("/agent-states", s.stateH.GetAgentStatesSummary)
	api.Get("/agent-state/metrics", s.stateH.GetStateMetrics)

	// Stripe payment routes (protected)
	stripe := api.Group("/stripe").Use(s.authH.RequireAuth())
	stripe.Post("/checkout", s.stripeH.CreateCheckoutSession)
	stripe.Post("/portal", s.stripeH.CreatePortalSession)
	stripe.Get("/subscription", s.stripeH.GetSubscriptionStatus)
	stripe.Get("/payments", s.stripeH.GetPaymentHistory)
	stripe.Get("/pricing", s.stripeH.GetPricing)

	// Stripe webhook (public - verified by signature)
	s.app.Post("/api/stripe/webhook", s.stripeH.HandleWebhook)

	// Analytics routes - Conversion Funnel Tracking
	// Public tracking endpoints (no auth required for anonymous tracking)
	analyticsPublic := api.Group("/analytics")
	analyticsPublic.Post("/track", s.analyticsH.TrackEvent)
	analyticsPublic.Post("/track/pricing-view", s.analyticsH.TrackPricingView)
	analyticsPublic.Post("/track/checkout-click", s.analyticsH.TrackCheckoutClick)
	analyticsPublic.Post("/track/checkout-cancel", s.analyticsH.TrackCheckoutCancel)

	// Protected analytics endpoints (require authentication)
	analyticsProtected := api.Group("/analytics").Use(s.authH.RequireAuth())
	analyticsProtected.Post("/track/checkout-start", s.analyticsH.TrackCheckoutStart)
	analyticsProtected.Post("/track/checkout-success", s.analyticsH.TrackCheckoutSuccess)
	analyticsProtected.Get("/funnel", s.analyticsH.GetFunnel)
	analyticsProtected.Get("/funnel/summary", s.analyticsH.GetFunnelSummary)
	analyticsProtected.Get("/funnel/daily", s.analyticsH.GetDailyStats)
	analyticsProtected.Get("/dashboard", s.analyticsH.GetDashboard)
	analyticsProtected.Get("/session/:sessionId", s.analyticsH.GetSessionEvents)

	// Newsletter subscription routes (public)
	newsletter := api.Group("/newsletter")
	newsletter.Post("/subscribe", s.newsletterH.Subscribe)
	newsletter.Post("/unsubscribe", s.newsletterH.Unsubscribe)
	newsletter.Get("/stats", s.newsletterH.GetStats)
	newsletter.Get("/recent", s.newsletterH.GetRecentSubscribers)
	newsletter.Get("/export", s.newsletterH.ExportSubscribers)

	// Beta signup routes (public for signup, protected for admin)
	beta := api.Group("/beta-signup")
	beta.Post("/", s.betaH.Signup)
	beta.Get("/", s.betaH.List)
	beta.Get("/stats", s.betaH.GetStats)
	beta.Get("/:id", s.betaH.Get)
	beta.Patch("/:id/status", s.betaH.UpdateStatus)
	beta.Delete("/:id", s.betaH.Delete)

	// Beta feedback routes (public for submission, protected for admin)
	feedback := api.Group("/beta-feedback")
	feedback.Post("/", s.betaFeedbackH.Submit)
	feedback.Get("/", s.betaFeedbackH.List)
	feedback.Get("/stats", s.betaFeedbackH.GetStats)
	feedback.Get("/:id", s.betaFeedbackH.Get)
	feedback.Patch("/:id/status", s.betaFeedbackH.UpdateStatus)
	feedback.Delete("/:id", s.betaFeedbackH.Delete)
}
