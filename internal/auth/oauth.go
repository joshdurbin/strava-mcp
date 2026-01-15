package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

const (
	authURL     = "https://www.strava.com/oauth/authorize"
	tokenURL    = "https://www.strava.com/oauth/token"
	redirectURI = "http://localhost:8089/callback"
	scopes      = "activity:read_all"
)

// StravaOAuthConfig returns an OAuth2 config for Strava
func StravaOAuthConfig(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      []string{scopes},
	}
}

// TokenResponse represents the OAuth token response from Strava
// We keep this for compatibility with existing storage
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// TokenFromOAuth2 converts an oauth2.Token to our TokenResponse
func TokenFromOAuth2(token *oauth2.Token) *TokenResponse {
	return &TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry.Unix(),
		TokenType:    token.TokenType,
	}
}

// ToOAuth2Token converts our TokenResponse to an oauth2.Token
func (t *TokenResponse) ToOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       time.Unix(t.ExpiresAt, 0),
		TokenType:    t.TokenType,
	}
}

// Authenticate performs the OAuth flow and returns tokens
func Authenticate(ctx context.Context, clientID, clientSecret string) (*TokenResponse, error) {
	config := StravaOAuthConfig(clientID, clientSecret)

	// Create channel to receive the authorization code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Create a new mux to avoid conflicts with default mux
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8089",
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "no authorization code received"
			}
			http.Error(w, errMsg, http.StatusBadRequest)
			errChan <- fmt.Errorf("authorization failed: %s", errMsg)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>Authorization successful!</h1><p>You can close this window.</p></body></html>`)
		codeChan <- code
	})

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Generate authorization URL with CSRF state
	state := "strava-mcp-auth"
	authURL := config.AuthCodeURL(state, oauth2.SetAuthURLParam("approval_prompt", "force"))

	fmt.Println("Opening browser for Strava authorization...")
	fmt.Printf("If browser doesn't open, visit: %s\n\n", authURL)

	// Open browser using pkg/browser
	if err := browser.OpenURL(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
	}

	// Wait for authorization code or error
	var code string
	select {
	case code = <-codeChan:
		// Success
	case err := <-errChan:
		server.Shutdown(ctx)
		return nil, err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return nil, fmt.Errorf("authorization timeout")
	}

	// Shutdown callback server
	server.Shutdown(ctx)

	// Exchange code for tokens using oauth2 library
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	return TokenFromOAuth2(token), nil
}

// RefreshAccessToken refreshes an expired access token using oauth2 library
func RefreshAccessToken(clientID, clientSecret, refreshToken string) (*TokenResponse, error) {
	config := StravaOAuthConfig(clientID, clientSecret)

	// Create an expired token with the refresh token
	oldToken := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Hour), // Force refresh
	}

	// Use TokenSource to refresh
	ctx := context.Background()
	tokenSource := config.TokenSource(ctx, oldToken)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return TokenFromOAuth2(newToken), nil
}

// IsTokenExpired checks if the token is expired or will expire soon
func IsTokenExpired(expiresAt int64) bool {
	// Consider expired if less than 5 minutes remaining
	return time.Now().Unix() > (expiresAt - 300)
}
