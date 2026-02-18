package cloud

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"
)

const browserOAuthTimeout = 120 * time.Second
const frontendURL = "https://app.mur.run"

const successHTML = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>mur - Login Successful</title></head>
<body style="font-family: -apple-system, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f8f9fa;">
<div style="text-align: center;">
<h1 style="color: #22c55e;">✓ Login Successful</h1>
<p>You can close this tab and return to the terminal.</p>
</div>
</body>
</html>`

const errorHTML = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>mur - Login Failed</title></head>
<body style="font-family: -apple-system, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f8f9fa;">
<div style="text-align: center;">
<h1 style="color: #ef4444;">✗ Login Failed</h1>
<p>%s</p>
<p>Please try again in the terminal.</p>
</div>
</body>
</html>`

// BrowserOAuthLogin performs OAuth login by opening a browser and receiving the callback.
func BrowserOAuthLogin(client *Client) error {
	// Generate random state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Start listener on random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// Channel to receive result
	type result struct {
		err error
	}
	done := make(chan result, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Validate state — don't terminate on mismatch, just reject this request
		if r.URL.Query().Get("state") != state {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, errorHTML, "Invalid state parameter (possible CSRF attack)")
			return
		}

		// Check for error from server
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, errorHTML, errMsg)
			done <- result{fmt.Errorf("oauth error: %s", errMsg)}
			return
		}

		var authData *AuthData

		if token := r.URL.Query().Get("token"); token != "" {
			// Email login: tokens provided directly
			authData = &AuthData{
				AccessToken:  token,
				RefreshToken: r.URL.Query().Get("refresh_token"),
				ExpiresAt:    time.Now().Add(1 * time.Hour),
			}
		} else if code := r.URL.Query().Get("code"); code != "" {
			// OAuth login: exchange code for tokens
			provider := r.URL.Query().Get("provider")
			if provider == "" {
				provider = "github" // default for backward compat
			}
			authResp, err := client.ExchangeOAuthCode(code, provider)
			if err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(w, errorHTML, "Failed to exchange authorization code")
				done <- result{fmt.Errorf("token exchange failed: %w", err)}
				return
			}
			authData = &AuthData{
				AccessToken:  authResp.AccessToken,
				RefreshToken: authResp.RefreshToken,
				ExpiresAt:    time.Now().Add(1 * time.Hour),
				User:         authResp.User,
			}
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, errorHTML, "No authorization code or token received")
			done <- result{fmt.Errorf("no code or token received")}
			return
		}

		if err := client.AuthStore().Save(authData); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, errorHTML, "Failed to save credentials")
			done <- result{fmt.Errorf("failed to save auth: %w", err)}
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, successHTML)
		done <- result{nil}
	})

	// Start server in background
	go func() {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			done <- result{fmt.Errorf("callback server error: %w", err)}
		}
	}()

	// Build OAuth URL and open browser
	oauthURL := fmt.Sprintf("%s/auth/cli-login?port=%d&state=%s",
		frontendURL, port, state)

	fmt.Println("Opening browser for authentication...")
	fmt.Println()

	if err := OpenURL(oauthURL); err != nil {
		srv.Close()
		return fmt.Errorf("failed to open browser: %w", err)
	}

	fmt.Printf("If the browser didn't open, visit:\n  %s\n\n", oauthURL)
	fmt.Println("Waiting for authentication...")

	// Wait for callback or timeout
	ctx, cancel := context.WithTimeout(context.Background(), browserOAuthTimeout)
	defer cancel()

	var loginErr error
	select {
	case res := <-done:
		loginErr = res.err
	case <-ctx.Done():
		loginErr = fmt.Errorf("authentication timed out after %s", browserOAuthTimeout)
	}

	// Shutdown server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)

	return loginErr
}
