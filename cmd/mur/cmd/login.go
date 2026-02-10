package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/mur-run/mur-core/internal/cloud"
	"github.com/mur-run/mur-core/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to mur-server",
	Long: `Authenticate with mur-server to enable team sync.

By default, uses device code flow (OAuth) - you'll authorize in your browser.
Use --password to login with email/password instead.

Examples:
  mur login                           # OAuth login (recommended)
  mur login --password                # Email/password login
  mur login --email user@example.com --password`,
	RunE: func(cmd *cobra.Command, args []string) error {
		usePassword, _ := cmd.Flags().GetBool("password")
		email, _ := cmd.Flags().GetString("email")
		serverURL, _ := cmd.Flags().GetString("server")

		// Get server URL from config if not specified
		if serverURL == "" {
			cfg, err := config.Load()
			if err == nil && cfg.Server.URL != "" {
				serverURL = cfg.Server.URL
			}
		}

		client, err := cloud.NewClient(serverURL)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Use device code flow by default
		if !usePassword && email == "" {
			return deviceCodeLogin(client)
		}

		// Fall back to email/password login
		return passwordLogin(client, email)
	},
}

func deviceCodeLogin(client *cloud.Client) error {
	fmt.Println("Starting device authorization...")
	fmt.Println()

	// Request device code
	codeResp, err := client.RequestDeviceCode()
	if err != nil {
		return fmt.Errorf("failed to start device authorization: %w", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  Open: %s\n", codeResp.VerificationURI)
	fmt.Println()
	fmt.Printf("  Enter code: %s\n", codeResp.UserCode)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Auto-open browser
	openBrowser(codeResp.VerificationURI)

	fmt.Println("Waiting for authorization...")

	// Poll for token
	pollInterval := time.Duration(codeResp.Interval) * time.Second
	if pollInterval < time.Second {
		pollInterval = 5 * time.Second
	}

	expiresAt := time.Now().Add(time.Duration(codeResp.ExpiresIn) * time.Second)

	for time.Now().Before(expiresAt) {
		time.Sleep(pollInterval)

		tokenResp, err := client.PollDeviceToken(codeResp.DeviceCode)
		if err != nil {
			if strings.Contains(err.Error(), "authorization_pending") {
				fmt.Print(".")
				continue
			}
			if strings.Contains(err.Error(), "expired") {
				return fmt.Errorf("authorization expired, please try again")
			}
			return fmt.Errorf("authorization failed: %w", err)
		}

		// Success!
		fmt.Println()
		fmt.Println()

		// Get user info
		user, err := client.Me()
		if err != nil {
			fmt.Println("✓ Logged in successfully")
		} else {
			fmt.Printf("✓ Logged in as %s (%s)\n", user.Name, user.Email)
		}
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  mur cloud teams     — List your teams")
		fmt.Println("  mur cloud sync      — Sync patterns with server")

		_ = tokenResp // tokens are stored by client
		return nil
	}

	return fmt.Errorf("authorization timed out")
}

func passwordLogin(client *cloud.Client, email string) error {
	reader := bufio.NewReader(os.Stdin)

	if email == "" {
		fmt.Print("Email: ")
		email, _ = reader.ReadString('\n')
		email = strings.TrimSpace(email)
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	if password == "" {
		return fmt.Errorf("password is required")
	}

	fmt.Println("Logging in...")

	resp, err := client.Login(email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println("")
	fmt.Printf("✓ Logged in as %s (%s)\n", resp.User.Name, resp.User.Email)
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  mur cloud teams     — List your teams")
	fmt.Println("  mur cloud sync      — Sync patterns with server")

	return nil
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from mur-server",
	Long:  `Clear stored credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cloud.NewClient("")
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		if err := client.Logout(); err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}

		fmt.Println("✓ Logged out")
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user",
	Long:  `Display the currently logged in user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")

		client, err := cloud.NewClient(serverURL)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		if !client.AuthStore().IsLoggedIn() {
			fmt.Println("Not logged in")
			fmt.Println("")
			fmt.Println("Run 'mur login' to authenticate")
			return nil
		}

		user, err := client.Me()
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		fmt.Printf("Logged in as %s (%s)\n", user.Name, user.Email)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)

	loginCmd.Flags().String("email", "", "Email address (for password login)")
	loginCmd.Flags().Bool("password", false, "Use email/password login instead of OAuth")
	loginCmd.Flags().String("server", "", "Server URL (default: https://api.mur.run)")
	whoamiCmd.Flags().String("server", "", "Server URL")
}
