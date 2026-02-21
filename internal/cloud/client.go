package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	ExpiresIn    int    `json:"expires_in"`
}

// Login authenticates with the server
func (c *Client) Login(email, password string) (*AuthResponse, error) {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp AuthResponse
	if err := c.post("/api/v1/core/auth/login", req, &resp); err != nil {
		return nil, err
	}

	// Use server-provided expiry, fallback to 365 days
	expiry := 365 * 24 * time.Hour
	if resp.ExpiresIn > 0 {
		expiry = time.Duration(resp.ExpiresIn) * time.Second
	}
	authData := &AuthData{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(expiry),
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
	if err := c.post("/api/v1/core/auth/refresh", req, &resp); err != nil {
		return err
	}

	// Use server-provided expiry, fallback to 365 days
	expiry := 365 * 24 * time.Hour
	if resp.ExpiresIn > 0 {
		expiry = time.Duration(resp.ExpiresIn) * time.Second
	}
	authData := &AuthData{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(expiry),
		User:         resp.User,
	}
	return c.authStore.Save(authData)
}

// Me returns the current user
func (c *Client) Me() (*User, error) {
	var user User
	if err := c.get("/api/v1/core/auth/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Logout clears stored credentials
func (c *Client) Logout() error {
	return c.authStore.Clear()
}

// LoginWithAPIKey authenticates using an API key
func (c *Client) LoginWithAPIKey(apiKey string) error {
	// Store the API key
	authData := &AuthData{
		APIKey: apiKey,
	}
	if err := c.authStore.Save(authData); err != nil {
		return err
	}

	// Verify it works by calling /me
	user, err := c.Me()
	if err != nil {
		// Clear the invalid key
		_ = c.authStore.Clear()
		return err
	}

	// Update with user info
	authData.User = user
	return c.authStore.Save(authData)
}

// DeviceCodeResponse represents device code response
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceTokenResponse represents device token response
type DeviceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error,omitempty"`
}

// RequestDeviceCode starts device authorization flow
func (c *Client) RequestDeviceCode() (*DeviceCodeResponse, error) {
	var resp DeviceCodeResponse
	if err := c.post("/api/v1/core/auth/device/code", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PollDeviceToken polls for device token
func (c *Client) PollDeviceToken(deviceCode string) (*DeviceTokenResponse, error) {
	req := map[string]string{"device_code": deviceCode}

	var resp DeviceTokenResponse
	if err := c.postRaw("/api/v1/core/auth/device/token", req, &resp); err != nil {
		// Check for expected errors
		if resp.Error != "" {
			return nil, fmt.Errorf("%s", resp.Error)
		}
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	// Save tokens (use server-provided expiry, fallback to 365 days)
	expiry := 365 * 24 * time.Hour
	if resp.ExpiresIn > 0 {
		expiry = time.Duration(resp.ExpiresIn) * time.Second
	}
	authData := &AuthData{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(expiry),
	}
	if err := c.authStore.Save(authData); err != nil {
		return nil, fmt.Errorf("failed to save auth: %w", err)
	}

	return &resp, nil
}

// ExchangeOAuthCode exchanges an OAuth code for a mur access token.
func (c *Client) ExchangeOAuthCode(code, provider string) (*AuthResponse, error) {
	req := map[string]string{"code": code, "provider": provider}
	var resp AuthResponse
	if err := c.post("/api/v1/core/auth/oauth/callback", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// postRaw is like post but doesn't fail on non-2xx if response is valid JSON
func (c *Client) postRaw(path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add device headers
	if c.deviceInfo != nil {
		req.Header.Set("X-Device-ID", c.deviceInfo.DeviceID)
		req.Header.Set("X-Device-Name", c.deviceInfo.DeviceName)
		req.Header.Set("X-Device-OS", c.deviceInfo.OS)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Always try to decode response
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

// Team represents a team
type Team struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	OwnerID            string    `json:"owner_id"`
	Plan               string    `json:"plan"`
	Role               string    `json:"role,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	SubscriptionStatus string    `json:"subscription_status,omitempty"`
	IsActive           bool      `json:"is_active"`
	CanSync            bool      `json:"can_sync"`
	CanInvite          bool      `json:"can_invite"`
}

// TeamsResponse represents teams list response
type TeamsResponse struct {
	Teams []Team `json:"teams"`
}

// ListTeams returns user's teams
func (c *Client) ListTeams() ([]Team, error) {
	var resp TeamsResponse
	if err := c.get("/api/v1/core/teams", &resp); err != nil {
		return nil, err
	}
	return resp.Teams, nil
}

// ResolveTeamID resolves a team slug or ID to a team ID
// If input looks like a UUID, returns it as-is; otherwise looks up by slug
func (c *Client) ResolveTeamID(slugOrID string) (string, error) {
	// Check if it's already a UUID
	if len(slugOrID) == 36 && strings.Count(slugOrID, "-") == 4 {
		return slugOrID, nil
	}

	// Look up by slug
	teams, err := c.ListTeams()
	if err != nil {
		return "", err
	}

	for _, t := range teams {
		if t.Slug == slugOrID {
			return t.ID, nil
		}
	}

	return "", fmt.Errorf("team '%s' not found", slugOrID)
}

// GetTeamBySlug returns team details by slug, including subscription status
func (c *Client) GetTeamBySlug(slug string) (*Team, error) {
	teams, err := c.ListTeams()
	if err != nil {
		return nil, err
	}

	for _, t := range teams {
		if t.Slug == slug {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("team '%s' not found", slug)
}

// CreateTeam creates a new team
func (c *Client) CreateTeam(name string) (*Team, error) {
	req := map[string]string{"name": name}
	var team Team
	if err := c.post("/api/v1/core/teams", req, &team); err != nil {
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
	path := fmt.Sprintf("/api/v1/core/teams/%s/sync/status?version=%d", teamID, version)
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
	path := fmt.Sprintf("/api/v1/core/teams/%s/sync/pull?since=%d", teamID, sinceVersion)
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
	ForceLocal  bool         `json:"force_local,omitempty"`
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
	path := fmt.Sprintf("/api/v1/core/teams/%s/sync/push", teamID)
	if err := c.post(path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// === Device Methods ===

// ListDevices returns all devices for the current user
func (c *Client) ListDevices() (*DeviceListResponse, error) {
	var resp DeviceListResponse
	if err := c.get("/api/v1/core/devices", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// LogoutDevice force-logs out a device
func (c *Client) LogoutDevice(deviceID string) error {
	return c.delete(fmt.Sprintf("/api/v1/core/devices/%s", deviceID))
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
	path := fmt.Sprintf("/api/v1/core/community/patterns/popular?limit=%d", limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCommunityRecent returns recent community patterns
func (c *Client) GetCommunityRecent(limit int) (*CommunityListResponse, error) {
	var resp CommunityListResponse
	path := fmt.Sprintf("/api/v1/core/community/patterns/recent?limit=%d", limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCommunityFeatured returns featured community patterns
func (c *Client) GetCommunityFeatured(limit int) (*CommunityListResponse, error) {
	var resp CommunityListResponse
	path := fmt.Sprintf("/api/v1/core/community/patterns/featured?limit=%d", limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UserProfile represents a user's public profile
type UserProfile struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Login        string             `json:"login,omitempty"`
	Bio          string             `json:"bio,omitempty"`
	Website      string             `json:"website,omitempty"`
	GitHub       string             `json:"github,omitempty"`
	Twitter      string             `json:"twitter,omitempty"`
	Plan         string             `json:"plan"`
	PatternCount int                `json:"pattern_count"`
	TotalCopies  int                `json:"total_copies"`
	TotalStars   int                `json:"total_stars"`
	Patterns     []CommunityPattern `json:"patterns"`
	JoinedAt     string             `json:"joined_at"`
}

// GetUserProfile returns a user's public profile
func (c *Client) GetUserProfile(login string) (*UserProfile, error) {
	var resp UserProfile
	path := fmt.Sprintf("/api/v1/core/community/users/%s", login)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Collection represents a pattern collection
type Collection struct {
	ID          string `json:"id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	CopyCount   int    `json:"copy_count"`
	CreatedAt   string `json:"created_at"`
}

// CollectionPattern represents a pattern in a collection
type CollectionPattern struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CopyCount   int    `json:"copy_count"`
}

// ListCollections returns public collections
func (c *Client) ListCollections(limit int) ([]Collection, error) {
	var resp struct {
		Collections []Collection `json:"collections"`
	}
	path := fmt.Sprintf("/api/v1/core/community/collections?limit=%d", limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Collections, nil
}

// GetCollection returns a collection with its patterns
func (c *Client) GetCollection(id string) (*Collection, []CollectionPattern, error) {
	var resp struct {
		Collection Collection          `json:"collection"`
		Patterns   []CollectionPattern `json:"patterns"`
	}
	path := fmt.Sprintf("/api/v1/core/community/collections/%s", id)
	if err := c.get(path, &resp); err != nil {
		return nil, nil, err
	}
	return &resp.Collection, resp.Patterns, nil
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(name, description, visibility string) (*Collection, error) {
	req := map[string]string{
		"name":        name,
		"description": description,
		"visibility":  visibility,
	}
	var resp Collection
	if err := c.post("/api/v1/core/community/collections", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchCommunity searches community patterns
func (c *Client) SearchCommunity(query string, limit int) (*CommunityListResponse, error) {
	return c.SearchCommunityWithTech(query, nil, limit)
}

// SearchCommunityWithTech searches community patterns with tech stack filter
func (c *Client) SearchCommunityWithTech(query string, techStack []string, limit int) (*CommunityListResponse, error) {
	var resp CommunityListResponse
	path := fmt.Sprintf("/api/v1/core/community/patterns/search?q=%s&limit=%d", url.QueryEscape(query), limit)

	// Add tech stack filter
	if len(techStack) > 0 {
		path += "&tech=" + strings.Join(techStack, ",")
	}

	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CommunityPatternDetail represents full pattern details from community
type CommunityPatternDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	AuthorName  string `json:"author_name"`
	AuthorLogin string `json:"author_login,omitempty"`
	CopyCount   int    `json:"copy_count"`
	ViewCount   int    `json:"view_count"`
	StarCount   int    `json:"star_count"`
}

// GetCommunityPattern gets full details of a community pattern
func (c *Client) GetCommunityPattern(id string) (*CommunityPatternDetail, error) {
	var resp CommunityPatternDetail
	path := fmt.Sprintf("/api/v1/core/community/patterns/%s", id)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CopyPattern copies a community pattern to user's team
func (c *Client) CopyPattern(patternID, teamID string) (*Pattern, error) {
	req := map[string]string{"team_id": teamID}
	var pattern Pattern
	path := fmt.Sprintf("/api/v1/core/community/patterns/%s/copy", patternID)
	if err := c.post(path, req, &pattern); err != nil {
		return nil, err
	}
	return &pattern, nil
}

// TeamPattern represents a pattern from a team
type TeamPattern struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Version     int64  `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TeamPatternsResponse is the response from listing team patterns
type TeamPatternsResponse struct {
	Patterns []TeamPattern `json:"patterns"`
	Total    int           `json:"total"`
}

// ListTeamPatterns lists patterns in a team
func (c *Client) ListTeamPatterns(teamSlug string, limit, offset int) ([]TeamPattern, int, error) {
	var resp TeamPatternsResponse
	path := fmt.Sprintf("/api/v1/core/teams/%s/patterns?limit=%d&offset=%d", teamSlug, limit, offset)
	if err := c.get(path, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Patterns, resp.Total, nil
}

// SharePatternRequest represents a request to share a pattern
type SharePatternRequest struct {
	PatternID   string   `json:"pattern_id"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
}

// SharePattern submits a pattern to community for review
func (c *Client) SharePattern(req *SharePatternRequest) error {
	path := fmt.Sprintf("/api/v1/core/community/patterns/%s/submit", req.PatternID)
	return c.post(path, req, nil)
}

// ShareLocalPatternRequest represents a request to share a local pattern directly
type ShareLocalPatternRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Domain      string   `json:"domain,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ShareLocalPatternResponse is the response when sharing a local pattern
type ShareLocalPatternResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"` // pending, approved, rejected
	Message string `json:"message,omitempty"`
}

// ShareLocalPattern uploads a local pattern directly to community
// Automatically translates non-English content before sharing
func (c *Client) ShareLocalPattern(req *ShareLocalPatternRequest) (*ShareLocalPatternResponse, error) {
	var resp ShareLocalPatternResponse
	if err := c.post("/api/v1/core/community/patterns/share", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TranslatePatternRequest represents a request to translate pattern content
type TranslatePatternRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// TranslatePatternResponse represents translated pattern content
type TranslatePatternResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// TranslatePattern translates pattern content to English using the server's LLM
func (c *Client) TranslatePattern(req *TranslatePatternRequest) (*TranslatePatternResponse, error) {
	var resp TranslatePatternResponse
	if err := c.post("/api/v1/core/community/translate", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
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
	if err := c.get("/api/v1/core/referral/stats", &stats); err != nil {
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
	if c.authStore.NeedsRefresh() && !strings.HasPrefix(path, "/api/v1/core/auth/") {
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
	token := c.authStore.GetToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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
