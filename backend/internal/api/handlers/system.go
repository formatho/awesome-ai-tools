package handlers

import (
	"runtime"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// SystemHandler handles system-related requests.
type SystemHandler struct {
	agentSvc *services.AgentService
	todoSvc  *services.TODOService
	cronSvc  *services.CronService
	started  time.Time
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(agentSvc *services.AgentService, todoSvc *services.TODOService, cronSvc *services.CronService) *SystemHandler {
	return &SystemHandler{
		agentSvc: agentSvc,
		todoSvc:  todoSvc,
		cronSvc:  cronSvc,
		started:  time.Now().UTC(),
	}
}

// Status returns the system status.
func (h *SystemHandler) Status(c *fiber.Ctx) error {
	// Get counts
	agents, _ := h.agentSvc.List()
	todos, _ := h.todoSvc.List()
	crons, _ := h.cronSvc.List()

	// Count by status
	agentCounts := make(map[string]int)
	for _, a := range agents {
		agentCounts[string(a.Status)]++
	}

	todoCounts := make(map[string]int)
	for _, t := range todos {
		todoCounts[string(t.Status)]++
	}

	cronCounts := make(map[string]int)
	for _, cr := range crons {
		cronCounts[string(cr.Status)]++
	}

	// Get memory stats
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return c.JSON(fiber.Map{
		"uptime_seconds": time.Since(h.started).Seconds(),
		"started_at":     h.started,
		"version":        "1.0.0",
		"counts": fiber.Map{
			"agents": len(agents),
			"todos":  len(todos),
			"crons":  len(crons),
		},
		"agents": fiber.Map{
			"by_status": agentCounts,
		},
		"todos": fiber.Map{
			"by_status": todoCounts,
		},
		"crons": fiber.Map{
			"by_status": cronCounts,
		},
		"resources": fiber.Map{
			"goroutines":   runtime.NumGoroutine(),
			"go_version":   runtime.Version(),
			"memory_mb":    mem.Alloc / 1024 / 1024,
			"sys_memory_mb": mem.Sys / 1024 / 1024,
			"num_cpu":      runtime.NumCPU(),
		},
	})
}

// Health returns a simple health check.
func (h *SystemHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}
