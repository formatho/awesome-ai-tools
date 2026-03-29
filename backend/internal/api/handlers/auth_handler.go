package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	
	// Proper JWT generation using base64 URL encoding
	// Format: header.payload.signature
	// Header (fixed for HS256)
	header := base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))
	
	// Payload with user claims and expiration
	payload := base64URLEncode([]byte(fmt.Sprintf(`{
		"user_id": %d,
		"email": "%s",
		"role": "%s",
		"exp": %d,
		"iat": %d,
		"iss": "agent-orchestrator"
	}`, userID, email, role, expiresAt.Unix(), time.Now().Unix())))
	
	// Create unsigned token for signature
	unsignedToken := header + "." + payload
	
	// Simple HMAC signature (NOT SECURE for production - use proper JWT library)
	signature := base64URLEncode(hmac256([]byte(unsignedToken), m.signingKey))
	
	token := unsignedToken + "." + signature
	return token, expiresAt, nil
}

// ValidateToken validates a JWT token and returns user information
func (m *JWTManager) ValidateToken(token string) (int64, string, string, error) {
	// Parse JWT token: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, "", "", fmt.Errorf("invalid token format")
	}

	header, payload, signature := parts[0], parts[1], parts[2]
	
	// Decode payload
	payloadData, err := base64URLDecode(payload)
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid payload: %v", err)
	}
	
	// Verify signature (simplified - in production would use proper JWT validation)
	expectedSignature := base64URLEncode(hmac256([]byte(header+"."+payload), m.signingKey))
	if signature != expectedSignature {
		return 0, "", "", fmt.Errorf("invalid signature")
	}
	
	// Parse payload JSON
	var claims struct {
		UserID int64  `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
		Exp    int64  `json:"exp"`
		Iat    int64  `json:"iat"`
		Iss    string `json:"iss"`
	}
	
	if err := json.Unmarshal(payloadData, &claims); err != nil {
		return 0, "", "", fmt.Errorf("invalid claims: %v", err)
	}
	
	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return 0, "", "", fmt.Errorf("token expired")
	}
	
	return claims.UserID, claims.Email, claims.Role, nil
}

// base64URLEncode encodes data using base64 URL encoding
func base64URLEncode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	// Replace URL-unsafe characters and remove padding
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	encoded = strings.ReplaceAll(encoded, "=", "")
	return encoded
}

// base64URLDecode decodes base64 URL encoded data
func base64URLDecode(data string) ([]byte, error) {
	// Add padding back if needed
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}
	
	// Replace URL-safe characters back to standard base64
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")
	
	return base64.StdEncoding.DecodeString(data)
}

// hmac256 creates HMAC-SHA256 signature
func hmac256(data, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
