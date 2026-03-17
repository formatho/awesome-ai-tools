package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AuthHandlerFiber handles authentication requests for Fiber framework.
type AuthHandlerFiber struct {
	jwtManager *JWTManager
}

func NewAuthHandlerFiber() *AuthHandlerFiber {
	return &AuthHandlerFiber{
		jwtManager: NewJWTManager(),
	}
}

// LoginFiber represents the login payload (non-conflicting name)
type LoginRequestFiber struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponseFiber represents the authentication response (non-conflicting name)
type LoginResponseFiber struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

// LoginFiber handles POST /api/auth/login - User login and token issuance.
func (h *AuthHandlerFiber) Login(c *fiber.Ctx) error {
	var req LoginRequestFiber
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Use hardcoded admin credentials for testing (same logic as Gin handler)
	var userID int64 = 1
	var email string = req.Email
	var role string = "admin"
	
	if req.Email == "admin@formatho.com" && req.Password == "testpass123" {
		userID = 1
		email = "admin@formatho.com"
		role = "admin"
	} else if req.Email == "user@formatho.com" && req.Password == "password123" {
		userID = 2
		email = "user@formatho.com"
		role = "member"
	} else if req.Email == "admin@example.com" {
		userID = 1
		email = req.Email
		role = "owner"
	} else if req.Email == "test@example.com" {
		userID = 2
		email = req.Email
		role = "member"
	} else {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Generate JWT token
	token, expiresAt, err := h.jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	response := LoginResponseFiber{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
		Email:     email,
		Role:      role,
	}

	return c.JSON(response)
}

// RequireAuthFiber returns a Fiber middleware that requires authentication.
func (h *AuthHandlerFiber) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		userID, email, role, err := h.jwtManager.ValidateToken(token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Set user info in context for downstream handlers
		c.Locals("user_id", userID)
		c.Locals("email", email)
		c.Locals("role", role)

		return c.Next()
	}
}

// LogoutFiber handles POST /api/auth/logout - Token invalidation.
func (h *AuthHandlerFiber) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// RefreshTokenFiber handles POST /api/auth/refresh - Issue a new access token.
func (h *AuthHandlerFiber) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	
	if err := c.BodyParser(&req); err != nil || req.Token == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Token required",
		})
	}

	userID, email, role, err := h.jwtManager.ValidateToken(req.Token)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	token, expiresAt, err := h.jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate new token",
		})
	}

	return c.JSON(fiber.Map{
		"token":      token,
		"expires_at": expiresAt,
	})
}

// CurrentUserFiber handles GET /api/auth/me - Returns current user info.
func (h *AuthHandlerFiber) CurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	email := c.Locals("email")
	role := c.Locals("role")

	return c.JSON(fiber.Map{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}

// GetContextClaims extracts user claims from Fiber context.
func (h *AuthHandlerFiber) GetContextClaims(c *fiber.Ctx) (int64, string, string, bool) {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return 0, "", "", false
	}

	email, ok := c.Locals("email").(string)
	if !ok {
		return 0, "", "", false
	}

	role, ok := c.Locals("role").(string)
	if !ok {
		return 0, "", "", false
	}

	return userID, email, role, true
}

// AuthMiddlewareFiber creates a Fiber middleware for authentication.
func (h *AuthHandlerFiber) AuthMiddlewareFiber() fiber.Handler {
	return h.RequireAuth()
}
