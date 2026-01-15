package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/joshdurbin/strava-mcp/internal/db"
)

// Storage handles auth data persistence using SQLite
type Storage struct {
	queries *db.Queries
	ctx     context.Context
}

// NewStorage creates a new Storage instance
func NewStorage(queries *db.Queries) *Storage {
	return &Storage{
		queries: queries,
		ctx:     context.Background(),
	}
}

// SaveTokens saves tokens to the database
func (s *Storage) SaveTokens(tokens *TokenResponse) error {
	// First try to get existing config to preserve client credentials
	_, err := s.queries.GetAuthConfig(s.ctx)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("checking existing config: %w", err)
	}

	if err == sql.ErrNoRows {
		// No config exists, can't save tokens without client config
		return fmt.Errorf("no client config found: run 'strava-mcp auth login' first")
	}

	// Update just the tokens
	return s.queries.UpdateTokens(s.ctx, db.UpdateTokensParams{
		AccessToken:  sql.NullString{String: tokens.AccessToken, Valid: true},
		RefreshToken: sql.NullString{String: tokens.RefreshToken, Valid: true},
		ExpiresAt:    sql.NullInt64{Int64: tokens.ExpiresAt, Valid: true},
	})
}

// LoadTokens loads tokens from the database
func (s *Storage) LoadTokens() (*StoredTokens, error) {
	config, err := s.queries.GetAuthConfig(s.ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("not authenticated: run 'strava-mcp auth login' first")
		}
		return nil, fmt.Errorf("loading auth config: %w", err)
	}

	if !config.AccessToken.Valid {
		return nil, fmt.Errorf("not authenticated: run 'strava-mcp auth login' first")
	}

	return &StoredTokens{
		AccessToken:  config.AccessToken.String,
		RefreshToken: config.RefreshToken.String,
		ExpiresAt:    config.ExpiresAt.Int64,
	}, nil
}

// SaveClientConfig saves client credentials and optional tokens to the database
func (s *Storage) SaveClientConfig(clientID, clientSecret string) error {
	return s.queries.SaveAuthConfig(s.ctx, db.SaveAuthConfigParams{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
}

// SaveFullConfig saves client credentials and tokens together
func (s *Storage) SaveFullConfig(clientID, clientSecret string, tokens *TokenResponse) error {
	return s.queries.SaveAuthConfig(s.ctx, db.SaveAuthConfigParams{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  sql.NullString{String: tokens.AccessToken, Valid: true},
		RefreshToken: sql.NullString{String: tokens.RefreshToken, Valid: true},
		ExpiresAt:    sql.NullInt64{Int64: tokens.ExpiresAt, Valid: true},
	})
}

// LoadClientConfig loads client credentials from the database
func (s *Storage) LoadClientConfig() (*ClientConfig, error) {
	config, err := s.queries.GetAuthConfig(s.ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("client not configured: run 'strava-mcp auth login' first")
		}
		return nil, fmt.Errorf("loading auth config: %w", err)
	}

	return &ClientConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
	}, nil
}

// DeleteTokens removes the stored auth config from the database
func (s *Storage) DeleteTokens() error {
	return s.queries.DeleteAuthConfig(s.ctx)
}

// GetValidAccessToken returns a valid access token, refreshing if necessary
func (s *Storage) GetValidAccessToken() (string, error) {
	tokens, err := s.LoadTokens()
	if err != nil {
		return "", err
	}

	// Check if token is still valid
	if !IsTokenExpired(tokens.ExpiresAt) {
		return tokens.AccessToken, nil
	}

	// Token expired, need to refresh
	config, err := s.LoadClientConfig()
	if err != nil {
		return "", fmt.Errorf("loading client config for refresh: %w", err)
	}

	newTokens, err := RefreshAccessToken(config.ClientID, config.ClientSecret, tokens.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("refreshing token: %w", err)
	}

	// Save the new tokens
	if err := s.SaveTokens(newTokens); err != nil {
		return "", fmt.Errorf("saving refreshed tokens: %w", err)
	}

	return newTokens.AccessToken, nil
}

// StoredTokens represents the tokens stored in the database
type StoredTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

// ClientConfig represents the stored client credentials
type ClientConfig struct {
	ClientID     string
	ClientSecret string
}
