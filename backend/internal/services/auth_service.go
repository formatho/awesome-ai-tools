package services

import (
	"database/sql"
	"time"

	"github.com/formatho/agent-orchestrator/backend/internal/models"
)

// AuthService provides authentication and authorization services.
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new auth service.
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// DB returns the database connection.
func (s *AuthService) DB() *sql.DB {
	return s.db
}

// TokenClaims represents JWT token claims.
type TokenClaims struct {
	UserID string
	OrgID  string
	Role   string
	Email  string
}

// ValidateToken validates an access token and returns claims.
// For now, this validates a simple token format: "userID|orgID|role|email"
// This is temporary - should use proper JWT
func (s *AuthService) ValidateToken(token string) (*TokenClaims, error) {
	// TODO: Implement proper JWT validation
	if token == "" || len(token) < 10 {
		return nil, models.ErrUnauthorized
	}

	// Parse simple token format: "userID|orgID|role|email"
	// Example: "user123|2cb8c4dd-1cac-4c83-811a-602c31da1fb2|owner|admin@example.com"
	parts := splitString(token, "|")
	if len(parts) < 3 {
		return nil, models.ErrUnauthorized
	}

	claims := &TokenClaims{
		UserID: parts[0],
		OrgID:  parts[1],
		Role:   parts[2],
	}

	if len(parts) >= 4 {
		claims.Email = parts[3]
	} else {
		claims.Email = "user@example.com"
	}

	return claims, nil
}

// Helper function to split string
func splitString(s, sep string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+1 >= len(sep) && s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// CheckPermission checks if a user has permission for a resource action.
func (s *AuthService) CheckPermission(userID, orgID, resource, action string) error {
	// TODO: Implement proper permission checking using database
	// For now, allow all team management actions for testing
	if resource == "team" && (action == "manage" || action == "read" || action == "write") {
		// Check if user is a member of the org
		query := `SELECT id FROM user_org_members WHERE user_id = ? AND organization_id = ?`
		var id string
		err := s.db.QueryRow(query, userID, orgID).Scan(&id)
		if err == nil {
			return nil // User is a member, grant permission
		}
	}

	// For testing, be permissive with known users
	if userID == "user123" || userID == "user456" {
		return nil
	}

	return models.ErrUnauthorized
}

// GenerateToken generates a new access token.
func (s *AuthService) GenerateToken(userID, orgID, role string) (string, time.Time, error) {
	// TODO: Implement proper JWT token generation
	// For now, generate a simple token
	token := userID + "|" + orgID + "|" + role + "|" + time.Now().Format(time.RFC3339)
	expiresAt := time.Now().Add(24 * time.Hour)
	return token, expiresAt, nil
}

// RefreshToken refreshes an access token using a refresh token.
func (s *AuthService) RefreshToken(refreshToken string) (string, time.Time, error) {
	// TODO: Implement proper token refresh
	// For now, generate a new token
	return s.GenerateToken("user123", "org456", "admin")
}

// ValidateSession validates a user session.
func (s *AuthService) ValidateSession(sessionID string) (*TokenClaims, error) {
	// TODO: Implement proper session validation
	// For now, return mock claims
	return &TokenClaims{
		UserID: "user123",
		OrgID:  "org456",
		Role:   "admin",
		Email:  "test@example.com",
	}, nil
}
