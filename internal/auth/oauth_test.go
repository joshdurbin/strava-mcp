package auth

import (
	"testing"
	"time"
)

func TestIsTokenExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiresAt int64
		want      bool
	}{
		{
			name:      "expired in the past",
			expiresAt: time.Now().Add(-1 * time.Hour).Unix(),
			want:      true,
		},
		{
			name:      "expires in 1 minute (within 5 min threshold)",
			expiresAt: time.Now().Add(1 * time.Minute).Unix(),
			want:      true,
		},
		{
			name:      "expires in 4 minutes (within 5 min threshold)",
			expiresAt: time.Now().Add(4 * time.Minute).Unix(),
			want:      true,
		},
		{
			name:      "expires in 10 minutes (beyond threshold)",
			expiresAt: time.Now().Add(10 * time.Minute).Unix(),
			want:      false,
		},
		{
			name:      "expires in 1 hour",
			expiresAt: time.Now().Add(1 * time.Hour).Unix(),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsTokenExpired(tt.expiresAt); got != tt.want {
				t.Errorf("IsTokenExpired(%d) = %v, want %v", tt.expiresAt, got, tt.want)
			}
		})
	}
}

func TestTokenFromOAuth2(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(1 * time.Hour)
	oauth2Token := &TokenResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		ExpiresAt:    expiry.Unix(),
		TokenType:    "Bearer",
	}

	converted := oauth2Token.ToOAuth2Token()

	if converted.AccessToken != "access_token" {
		t.Errorf("expected access token 'access_token', got %q", converted.AccessToken)
	}

	if converted.RefreshToken != "refresh_token" {
		t.Errorf("expected refresh token 'refresh_token', got %q", converted.RefreshToken)
	}

	if converted.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got %q", converted.TokenType)
	}

	// Convert back
	backConverted := TokenFromOAuth2(converted)

	if backConverted.AccessToken != oauth2Token.AccessToken {
		t.Errorf("round-trip failed: access token mismatch")
	}

	if backConverted.RefreshToken != oauth2Token.RefreshToken {
		t.Errorf("round-trip failed: refresh token mismatch")
	}
}

func TestStravaOAuthConfig(t *testing.T) {
	t.Parallel()

	config := StravaOAuthConfig("test_client_id", "test_client_secret")

	if config.ClientID != "test_client_id" {
		t.Errorf("expected client_id 'test_client_id', got %q", config.ClientID)
	}

	if config.ClientSecret != "test_client_secret" {
		t.Errorf("expected client_secret 'test_client_secret', got %q", config.ClientSecret)
	}

	if config.Endpoint.AuthURL != "https://www.strava.com/oauth/authorize" {
		t.Errorf("unexpected auth URL: %q", config.Endpoint.AuthURL)
	}

	if config.Endpoint.TokenURL != "https://www.strava.com/oauth/token" {
		t.Errorf("unexpected token URL: %q", config.Endpoint.TokenURL)
	}

	if config.RedirectURL != "http://localhost:8089/callback" {
		t.Errorf("unexpected redirect URL: %q", config.RedirectURL)
	}
}
