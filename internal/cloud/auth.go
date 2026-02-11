package cloud

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuthStore manages authentication tokens
type AuthStore struct {
	path string
}

// AuthData represents stored auth data
type AuthData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user,omitempty"`
	APIKey       string    `json:"api_key,omitempty"` // API key for authentication (never expires)
}

// User represents a mur-server user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// NewAuthStore creates a new auth store
func NewAuthStore() (*AuthStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	murDir := filepath.Join(home, ".mur")
	if err := os.MkdirAll(murDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .mur dir: %w", err)
	}

	return &AuthStore{
		path: filepath.Join(murDir, "auth.json"),
	}, nil
}

// Save saves auth data
func (s *AuthStore) Save(data *AuthData) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	if err := os.WriteFile(s.path, b, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// Load loads auth data
func (s *AuthStore) Load() (*AuthData, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	var data AuthData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("failed to parse auth file: %w", err)
	}

	return &data, nil
}

// Clear removes auth data
func (s *AuthStore) Clear() error {
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove auth file: %w", err)
	}
	return nil
}

// IsLoggedIn checks if user is logged in with valid token or API key
func (s *AuthStore) IsLoggedIn() bool {
	data, err := s.Load()
	if err != nil || data == nil {
		return false
	}
	// API key never expires
	if data.APIKey != "" {
		return true
	}
	return data.AccessToken != "" && time.Now().Before(data.ExpiresAt)
}

// GetToken returns the token to use for authentication (API key or access token)
func (s *AuthStore) GetToken() string {
	data, err := s.Load()
	if err != nil || data == nil {
		return ""
	}
	// Prefer API key if set
	if data.APIKey != "" {
		return data.APIKey
	}
	return data.AccessToken
}

// NeedsRefresh checks if token needs refresh (expires in < 5 minutes)
func (s *AuthStore) NeedsRefresh() bool {
	data, err := s.Load()
	if err != nil || data == nil {
		return false
	}
	return time.Now().Add(5 * time.Minute).After(data.ExpiresAt)
}
