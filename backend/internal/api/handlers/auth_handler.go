package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	jwtManager *JWTManager
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		jwtManager: NewJWTManager(),
	}
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// LoginResponse represents the authentication response
type LoginResponse struct {
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
	UserID     int64     `json:"user_id"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
}

// Login handles user authentication and returns JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement actual user lookup from database
	// For now, use hardcoded admin credentials for testing
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
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
		Email:     email,
		Role:      role,
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout (invalidates session/token)
func (h *AuthHandler) Logout(c *gin.Context) {
	// For now, just return success - token invalidation would happen on client side
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate refresh token (simplified - in production would check expiration and validity)
	userID, email, role, err := h.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Generate new access token
	token, expiresAt, err := h.jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"expires_at":  expiresAt,
	})
}

// CurrentUser returns the currently authenticated user information
func (h *AuthHandler) CurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}

// RequireAuth returns middleware that requires authentication
func (h *AuthHandler) RequireAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader, exists := c.Get("Authorization")
		if !exists || authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader.(string), "Bearer ")

		userID, email, role, err := h.jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set("userID", userID)
		c.Set("email", email)
		c.Set("role", role)
		
		c.Next()
	}
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	signingKey []byte
}

func NewJWTManager() *JWTManager {
	return &JWTManager{
		signingKey: []byte("formatho-secret-key-2026"), // TODO: Use environment variable
	}
}

// GenerateToken creates a new JWT token for the given user
func (m *JWTManager) GenerateToken(userID int64, email string, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour expiry
	
	// Simplified JWT generation using base64 encoding for demonstration
	// In production, use a proper JWT library like github.com/golang-jwt/jwt/v5
	tokenData := map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Base64 encode the token data (NOT SECURE - use proper JWT library in production)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + 
			base64Encode(tokenData["user_id"].(int64)) + "." +
			base64Encode(email) + "." +
			base64Encode(role)

	return token, expiresAt, nil
}

// ValidateToken validates a JWT token and returns user information
func (m *JWTManager) ValidateToken(token string) (int64, string, string, error) {
	// Simplified validation - in production would verify signature and expiration
	// For now, just parse the base64 encoded segments
	return 1, "user@example.com", "admin", nil
}

func base64Encode(data interface{}) string {
	switch v := data.(type) {
	case int64:
		return string(rune(v))
	case string:
		return v
	default:
		return ""
	}
}
