package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
	"github.com/formatho/agent-orchestrator/backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

// BetaFeedbackHandler handles beta feedback-related HTTP requests
type BetaFeedbackHandler struct {
	feedbackSvc *services.BetaFeedbackService
	emailSvc    *services.EmailService
}

// NewBetaFeedbackHandler creates a new beta feedback handler
func NewBetaFeedbackHandler(feedbackSvc *services.BetaFeedbackService, emailSvc *services.EmailService) *BetaFeedbackHandler {
	return &BetaFeedbackHandler{
		feedbackSvc: feedbackSvc,
		emailSvc:    emailSvc,
	}
}

// SubmitFeedback handles POST /api/beta-feedback
func (h *BetaFeedbackHandler) SubmitFeedback(c *fiber.Ctx) error {
	// Parse form data
	feedback := &models.BetaFeedback{
		Type:              c.FormValue("type"),
		Rating:            parseInt(c.FormValue("rating")),
		Title:             c.FormValue("title"),
		Description:       c.FormValue("description"),
		Email:             c.FormValue("email"),
		Name:              c.FormValue("name"),
		Priority:          c.FormValue("priority"),
		Browser:           c.FormValue("browser"),
		StepsToReproduce:  c.FormValue("steps_to_reproduce"),
		ExpectedBehavior:  c.FormValue("expected_behavior"),
		ActualBehavior:    c.FormValue("actual_behavior"),
		BetaTesterID:      c.FormValue("beta_tester_id"),
		Status:           "new",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Validate required fields
	if feedback.Title == "" || feedback.Description == "" || feedback.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title, description, and email are required",
		})
	}

	// Handle file uploads
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := form.File["files"]
		for _, file := range files {
			// Save file and store path
			filePath := fmt.Sprintf("/uploads/feedback/%s_%s", time.Now().Format("20060102-150405"), file.Filename)
			if err := c.SaveFile(file, "."+filePath); err != nil {
				log.Printf("Error saving file: %v", err)
			} else {
				feedback.Attachments = append(feedback.Attachments, filePath)
			}
		}
	}

	// Save feedback to database
	created, err := h.feedbackSvc.Create(feedback)
	if err != nil {
		log.Printf("Error creating feedback: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to submit feedback",
		})
	}

	// Send email notification to team
	go h.sendNotificationEmail(created)

	// Create task in agent-todo system if it's a bug or feature request
	if feedback.Type == "bug" || feedback.Type == "feature" {
		go h.createAgentTodoTask(created)
	}

	log.Printf("Successfully received beta feedback: %s from %s", feedback.Title, feedback.Email)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":  true,
		"message":  "Feedback submitted successfully",
		"feedback": created,
	})
}

// ListFeedback handles GET /api/beta-feedback
func (h *BetaFeedbackHandler) ListFeedback(c *fiber.Ctx) error {
	status := c.Query("status")
	feedbackType := c.Query("type")
	priority := c.Query("priority")

	feedbacks, err := h.feedbackSvc.List(status, feedbackType, priority)
	if err != nil {
		log.Printf("Error listing feedback: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list feedback",
		})
	}

	return c.JSON(fiber.Map{
		"feedback": feedbacks,
		"count":    len(feedbacks),
	})
}

// GetFeedback handles GET /api/beta-feedback/:id
func (h *BetaFeedbackHandler) GetFeedback(c *fiber.Ctx) error {
	id := c.Params("id")
	
	feedback, err := h.feedbackSvc.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Feedback not found",
		})
	}

	return c.JSON(feedback)
}

// UpdateFeedbackStatus handles PUT /api/beta-feedback/:id/status
func (h *BetaFeedbackHandler) UpdateFeedbackStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var req struct {
		Status      string `json:"status"`
		Resolution  string `json:"resolution"`
		AssignedTo  string `json:"assigned_to"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	feedback, err := h.feedbackSvc.UpdateStatus(id, req.Status, req.Resolution, req.AssignedTo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update feedback",
		})
	}

	// Send status update email to submitter
	go h.sendStatusUpdateEmail(feedback)

	return c.JSON(feedback)
}

// GetFeedbackStats handles GET /api/beta-feedback/stats
func (h *BetaFeedbackHandler) GetFeedbackStats(c *fiber.Ctx) error {
	stats, err := h.feedbackSvc.GetStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get stats",
		})
	}

	return c.JSON(stats)
}

// Helper functions

func (h *BetaFeedbackHandler) sendNotificationEmail(feedback *models.BetaFeedback) {
	if h.emailSvc == nil {
		return
	}

	subject := fmt.Sprintf("[Beta Feedback] %s: %s", feedback.Type, feedback.Title)
	body := fmt.Sprintf(`
New beta feedback received:

Type: %s
Priority: %s
From: %s (%s)
Title: %s

Description:
%s

%s

View in dashboard: https://app.formatho.com/admin/feedback/%s
`, 
		feedback.Type,
		feedback.Priority,
		feedback.Name,
		feedback.Email,
		feedback.Title,
		feedback.Description,
		h.formatBugDetails(feedback),
		feedback.ID,
	)

	err := h.emailSvc.SendEmail("founders@formatho.com", subject, body)
	if err != nil {
		log.Printf("Failed to send notification email: %v", err)
	}
}

func (h *BetaFeedbackHandler) sendStatusUpdateEmail(feedback *models.BetaFeedback) {
	if h.emailSvc == nil {
		return
	}

	subject := fmt.Sprintf("Update on your feedback: %s", feedback.Title)
	body := fmt.Sprintf(`
Hi %s,

Your feedback "%s" has been updated.

Status: %s
%s

Thank you for helping us improve Formatho!

Best,
The Formatho Team
`, 
		feedback.Name,
		feedback.Title,
		feedback.Status,
		h.formatResolution(feedback),
	)

	err := h.emailSvc.SendEmail(feedback.Email, subject, body)
	if err != nil {
		log.Printf("Failed to send status update email: %v", err)
	}
}

func (h *BetaFeedbackHandler) createAgentTodoTask(feedback *models.BetaFeedback) {
	// Create task in agent-todo system
	taskData := map[string]interface{}{
		"title":       fmt.Sprintf("[Beta Feedback] %s", feedback.Title),
		"description": feedback.Description,
		"priority":    feedback.Priority,
		"source":      "beta_feedback",
		"source_id":   feedback.ID,
		"submitter":   feedback.Email,
	}

	jsonData, _ := json.Marshal(taskData)
	
	// Call agent-todo API
	resp, err := http.Post(
		"http://localhost:18765/api/tasks",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		log.Printf("Failed to create agent-todo task: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Created agent-todo task for feedback %s", feedback.ID)
}

func (h *BetaFeedbackHandler) formatBugDetails(feedback *models.BetaFeedback) string {
	if feedback.Type != "bug" {
		return ""
	}
	
	return fmt.Sprintf(`
Bug Details:
- Steps to Reproduce: %s
- Expected: %s
- Actual: %s
- Browser: %s
`, 
		feedback.StepsToReproduce,
		feedback.ExpectedBehavior,
		feedback.ActualBehavior,
		feedback.Browser,
	)
}

func (h *BetaFeedbackHandler) formatResolution(feedback *models.BetaFeedback) string {
	if feedback.Resolution == "" {
		return ""
	}
	
	return fmt.Sprintf("\nResolution:\n%s\n", feedback.Resolution)
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
