package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultServerURL = "https://api.mur.run"
)

// Client is the mur-server API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	authStore  *AuthStore
	deviceInfo *DeviceInfo
}

// NewClient creates a new API client
func NewClient(serverURL string) (*Client, error) {
	if serverURL == "" {
		serverURL = DefaultServerURL
	}

	authStore, err := NewAuthStore()
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:    serverURL,
		authStore:  authStore,
		deviceInfo: GetDeviceInfo(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// AuthStore returns the auth store
func (c *Client) AuthStore() *AuthStore {
	return c.authStore
}

// LoginRequest represents login input
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents auth response
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login authenticates with the server
func (c *Client) Login(email, password string) (*AuthResponse, error) {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp AuthResponse
	if err := c.post("/api/v1/auth/login", req, &resp); err != nil {
		return nil, err
	}

	// Save tokens (1 hour expiry, matching server)
	authData := &AuthData{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		User:         resp.User,
	}
	if err := c.authStore.Save(authData); err != nil {
		return nil, fmt.Errorf("failed to save auth: %w", err)
	}

	return &resp, nil
}

// Refresh refreshes the access token
func (c *Client) Refresh() error {
	auth, err := c.authStore.Load()
	if err != nil || auth == nil {
		return fmt.Errorf("not logged in")
	}

	req := map[string]string{
		"refresh_token": auth.RefreshToken,
	}

	var resp AuthResponse
	if err := c.post("/api/v1/auth/refresh", req, &resp); err != nil {
		return err
	}

	// Save new tokens (1 hour expiry, matching server)
	authData := &AuthData{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		User:         resp.User,
	}
	return c.authStore.Save(authData)
}

// Me returns the current user
func (c *Client) Me() (*User, error) {
	var user User
	if err := c.get("/api/v1/auth/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Logout clears stored credentials
func (c *Client) Logout() error {
	return c.authStore.Clear()
}

// Team represents a team
type Team struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"owner_id"`
	Plan      string    `json:"plan"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TeamsResponse represents teams list response
type TeamsResponse struct {
	Teams []Team `json:"teams"`
}

// ListTeams returns user's teams
func (c *Client) ListTeams() ([]Team, error) {
	var resp TeamsResponse
	if err := c.get("/api/v1/teams", &resp); err != nil {
		return nil, err
	}
	return resp.Teams, nil
}

// CreateTeam creates a new team
func (c *Client) CreateTeam(name string) (*Team, error) {
	req := map[string]string{"name": name}
	var team Team
	if err := c.post("/api/v1/teams", req, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

// SyncStatus represents sync status
type SyncStatus struct {
	ServerVersion int64 `json:"server_version"`
	HasUpdates    bool  `json:"has_updates"`
}

// GetSyncStatus returns sync status
func (c *Client) GetSyncStatus(teamID string, version int64) (*SyncStatus, error) {
	var status SyncStatus
	path := fmt.Sprintf("/api/v1/teams/%s/sync/status?version=%d", teamID, version)
	if err := c.get(path, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Pattern represents a pattern
type Pattern struct {
	ID          string         `json:"id"`
	TeamID      string         `json:"team_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Content     string         `json:"content"`
	Tags        map[string]any `json:"tags"`
	Applies     map[string]any `json:"applies"`
	Security    map[string]any `json:"security"`
	Learning    map[string]any `json:"learning"`
	Lifecycle   map[string]any `json:"lifecycle"`
	Version     int64          `json:"version"`
	Deleted     bool           `json:"deleted"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	// v1.1.0+ fields for pattern schema v2
	PatternVersion string `json:"pattern_version,omitempty"`
	SchemaVersion  int    `json:"schema_version,omitempty"`
	EmbeddingHash  string `json:"embedding_hash,omitempty"`
}

// PullResponse represents pull response
type PullResponse struct {
	Patterns []Pattern `json:"patterns"`
	Version  int64     `json:"version"`
}

// Pull pulls patterns since a version
func (c *Client) Pull(teamID string, sinceVersion int64) (*PullResponse, error) {
	var resp PullResponse
	path := fmt.Sprintf("/api/v1/teams/%s/sync/pull?since=%d", teamID, sinceVersion)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SyncChange represents a sync change
type SyncChange struct {
	Action  string   `json:"action"` // create, update, delete
	ID      string   `json:"id,omitempty"`
	Pattern *Pattern `json:"pattern,omitempty"`
}

// PushRequest represents push request
type PushRequest struct {
	BaseVersion int64        `json:"base_version"`
	Changes     []SyncChange `json:"changes"`
}

// Conflict represents a sync conflict
type Conflict struct {
	PatternID     string   `json:"pattern_id"`
	PatternName   string   `json:"pattern_name"`
	ServerVersion *Pattern `json:"server_version"`
	ClientVersion *Pattern `json:"client_version"`
}

// PushResponse represents push response
type PushResponse struct {
	OK        bool       `json:"ok"`
	Version   int64      `json:"version"`
	Conflicts []Conflict `json:"conflicts,omitempty"`
}

// Push pushes changes
func (c *Client) Push(teamID string, req PushRequest) (*PushResponse, error) {
	var resp PushResponse
	path := fmt.Sprintf("/api/v1/teams/%s/sync/push", teamID)
	if err := c.post(path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// === Device Methods ===

// ListDevices returns all devices for the current user
func (c *Client) ListDevices() (*DeviceListResponse, error) {
	var resp DeviceListResponse
	if err := c.get("/api/v1/devices", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// LogoutDevice force-logs out a device
func (c *Client) LogoutDevice(deviceID string) error {
	return c.delete(fmt.Sprintf("/api/v1/devices/%s", deviceID))
}

// === Community Methods ===

// CommunityPattern represents a pattern in the community
type CommunityPattern struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name"`
	AuthorLogin string `json:"author_login,omitempty"`
	CopyCount   int    `json:"copy_count"`
	ViewCount   int    `json:"view_count"`
}

// CommunityListResponse is the response from community endpoints
type CommunityListResponse struct {
	Patterns []CommunityPattern `json:"patterns"`
	Count    int                `json:"count"`
}

// GetCommunityPopular returns popular community patterns
func (c *Client) GetCommunityPopular(limit int) (*CommunityListResponse, error) {
	var resp CommunityListResponse
	path := fmt.Sprintf("/api/v1/community/patterns/popular?limit=%d", limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCommunityRecent returns recent community patterns
func (c *Client) GetCommunityRecent(limit int) (*CommunityListResponse, error) {
	var resp CommunityListResponse
	path := fmt.Sprintf("/api/v1/community/patterns/recent?limit=%d", limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchCommunity searches community patterns
func (c *Client) SearchCommunity(query string, limit int) (*CommunityListResponse, error) {
	var resp CommunityListResponse
	path := fmt.Sprintf("/api/v1/community/patterns/search?q=%s&limit=%d", query, limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CopyPattern copies a community pattern to user's team
func (c *Client) CopyPattern(patternID, teamID string) (*Pattern, error) {
	req := map[string]string{"team_id": teamID}
	var pattern Pattern
	path := fmt.Sprintf("/api/v1/community/patterns/%s/copy", patternID)
	if err := c.post(path, req, &pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}

// === Referral Methods ===

// ReferralStats represents referral statistics
type ReferralStats struct {
	ReferralCode   string `json:"referral_code"`
	ReferralLink   string `json:"referral_link"`
	TotalShared    int    `json:"total_shared"`
	TotalQualified int    `json:"total_qualified"`
	TotalRewarded  int    `json:"total_rewarded"`
	RewardsLeft    int    `json:"rewards_left"`
	DaysEarned     int    `json:"days_earned"`
}

// GetReferralStats returns referral statistics
func (c *Client) GetReferralStats() (*ReferralStats, error) {
	var stats ReferralStats
	if err := c.get("/api/v1/referral/stats", &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// HTTP helpers

func (c *Client) get(path string, result interface{}) error {
	return c.do("GET", path, nil, result)
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	return c.do("POST", path, body, result)
}

func (c *Client) delete(path string) error {
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	// Auto-refresh token if needed (but not for auth endpoints to avoid recursion)
	if c.authStore.NeedsRefresh() && !strings.HasPrefix(path, "/api/v1/auth/") {
		_ = c.Refresh() // Ignore refresh errors, request will fail if token invalid
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add device headers
	if c.deviceInfo != nil {
		req.Header.Set("X-Device-ID", c.deviceInfo.DeviceID)
		req.Header.Set("X-Device-Name", c.deviceInfo.DeviceName)
		req.Header.Set("X-Device-OS", c.deviceInfo.OS)
	}

	// Add auth header if logged in
	auth, _ := c.authStore.Load()
	if auth != nil && auth.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Check for device limit error (429)
		if resp.StatusCode == 429 {
			var deviceErr struct {
				Error   string   `json:"error"`
				Message string   `json:"message"`
				Limit   int      `json:"limit"`
				Active  []Device `json:"active"`
			}
			if json.Unmarshal(respBody, &deviceErr) == nil && deviceErr.Error == "device_limit_exceeded" {
				return &DeviceLimitError{
					Limit:   deviceErr.Limit,
					Active:  deviceErr.Active,
					Message: deviceErr.Message,
				}
			}
		}

		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
