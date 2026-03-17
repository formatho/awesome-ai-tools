// Package auth provides JWT-based authentication services.
package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken    = errors.New("invalid token")
	ErrExpiredToken    = errors.New("token has expired")
	ErrInvalidClaims   = errors.New("invalid claims")
	ErrTokenRefresh    = errors.New("failed to refresh token")
	ErrLogoutFailed    = errors.New("failed to logout")
)

// Claims represents JWT claims.
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	OrgID     string `json:"org_id,omitempty"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Config holds JWT configuration.
type Config struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

// DefaultConfig returns a default JWT configuration.
func DefaultConfig() *Config {
	return &Config{
		AccessSecret:  []byte("your-access-secret-key-change-in-production"),
		RefreshSecret: []byte("your-refresh-secret-key-change-in-production"),
		AccessExpire:  15 * time.Minute,
		RefreshExpire: 7 * 24 * time.Hour, // 7 days
	}
}

// TokenStore defines the interface for token storage.
type TokenStore interface {
	SaveToken(token string, userId string, expiresAt time.Time) error
	RevokeToken(token string) error
	IsTokenRevoked(token string) (bool, error)
	GetUserTokens(userId string) ([]string, error)
}

// InMemoryTokenStore is a simple in-memory token store for development.
type InMemoryTokenStore struct {
	mu         sync.RWMutex
	revoked    map[string]time.Time // token -> expiry time
	userTokens map[string][]string  // userId -> []token
}

// NewInMemoryTokenStore creates a new in-memory token store.
func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		revoked:    make(map[string]time.Time),
		userTokens: make(map[string][]string),
	}
}

// SaveToken stores a token for the user.
func (s *InMemoryTokenStore) SaveToken(token string, userId string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.userTokens[userId]; !ok {
		s.userTokens[userId] = []string{}
	}
	s.userTokens[userId] = append(s.userTokens[userId], token)
	return nil
}

// RevokeToken revokes a specific token.
func (s *InMemoryTokenStore) RevokeToken(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.revoked, token)
	for userId, tokens := range s.userTokens {
		for i, t := range tokens {
			if t == token {
				s.userTokens[userId] = append(tokens[:i], tokens[i+1:]...)
				break
			}
		}
	}
	return nil
}

// IsTokenRevoked checks if a token has been revoked.
func (s *InMemoryTokenStore) IsTokenRevoked(token string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if expiry, ok := s.revoked[token]; ok {
		return time.Now().Before(expiry), nil
	}
	return false, nil
}

// GetUserTokens returns all tokens for a user.
func (s *InMemoryTokenStore) GetUserTokens(userId string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tokens := s.userTokens[userId]
	result := make([]string, len(tokens))
	copy(result, tokens)
	return result, nil
}

// AuthService provides JWT authentication services.
type AuthService struct {
	config     *Config
	tokenStore TokenStore
	keyFunc    jwt.Keyfunc
}

// NewAuthService creates a new authentication service.
func NewAuthService(config *Config, store TokenStore) *AuthService {
	if config == nil {
		config = DefaultConfig()
	}
	if store == nil {
		store = NewInMemoryTokenStore()
	}

	return &AuthService{
		config:     config,
		tokenStore: store,
		keyFunc: func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.AccessSecret, nil
		},
	}
}

// Login authenticates a user and returns an access token.
func (s *AuthService) Login(email, password string) (string, *Claims, error) {
	// TODO: Implement actual authentication against user database
	// For now, return a mock token for development
	// In production, this should validate credentials against the database

	if email == "" || password == "" {
		return "", nil, ErrInvalidClaims
	}

	claims := &Claims{
		UserID:    "user-123", // TODO: Get from DB after auth
		Email:     email,
		Role:      "admin",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "agent-orchestrator",
			Subject:   email,
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.config.AccessSecret)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Store token for session management
	if err := s.tokenStore.SaveToken(token, claims.UserID, claims.ExpiresAt.Time); err != nil {
		return "", nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, claims, nil
}

// ValidateToken validates a JWT access token and returns its claims.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.AccessSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is revoked
		isRevoked, err := s.tokenStore.IsTokenRevoked(tokenString)
		if err != nil {
			return nil, fmt.Errorf("failed to check token revocation: %w", err)
		}
		if isRevoked {
			return nil, ErrInvalidToken
		}

		return claims, nil
	}

	return nil, ErrInvalidClaims
}

// ValidateRefreshToken validates a JWT refresh token.
func (s *AuthService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.RefreshSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if isRevoked, err := s.tokenStore.IsTokenRevoked(tokenString); err != nil || isRevoked {
			return nil, ErrInvalidToken
		}
		return claims, nil
	}

	return nil, ErrInvalidClaims
}

// RefreshToken generates a new access token using a valid refresh token.
func (s *AuthService) RefreshToken(oldRefreshToken string) (string, *Claims, error) {
	claims, err := s.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return "", nil, ErrInvalidClaims
	}

	newAccessToken := &Claims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		Role:      claims.Role,
		OrgID:     claims.OrgID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "agent-orchestrator",
			Subject:   claims.Email,
		},
	}

	newToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessToken).SignedString(s.config.AccessSecret)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create access token: %w", err)
	}

	if err := s.tokenStore.SaveToken(newToken, claims.UserID, newAccessToken.ExpiresAt.Time); err != nil {
		return "", nil, fmt.Errorf("failed to save token: %w", err)
	}

	return newToken, newAccessToken, nil
}

// Logout invalidates a user's session by revoking tokens.
func (s *AuthService) Logout(userId string, token string) error {
	// Revoke the specific token
	if err := s.tokenStore.RevokeToken(token); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	// TODO: In production, you might want to revoke all user tokens for security
	// tokens, _ := s.tokenStore.GetUserTokens(userId)
	// for _, t := range tokens {
	//     s.tokenStore.RevokeToken(t)
	// }

	return nil
}

// GenerateRefreshToken creates a new refresh token for a user.
func (s *AuthService) GenerateRefreshToken(userID, email, role string, orgID string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		OrgID:     orgID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "agent-orchestrator",
			Subject:   email,
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.config.RefreshSecret)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err := s.tokenStore.SaveToken(token, userID, claims.ExpiresAt.Time); err != nil {
		return "", fmt.Errorf("failed to save refresh token: %w", err)
	}

	return token, nil
}
