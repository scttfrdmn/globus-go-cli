// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package mocks

import (
	"context"
	"time"
)

// These are simplified mock types to reduce dependencies

// TokenIntrospection represents a token introspection response
type TokenIntrospection struct {
	Active      bool     `json:"active"`
	Scope       string   `json:"scope"`
	ClientID    string   `json:"client_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Subject     string   `json:"sub"`
	Exp         int64    `json:"exp"`
	IdentitySet []string `json:"identity_set"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope"`
	ExpiryTime   time.Time `json:"expiry_time"`
}

// MockAuthClient implements a mock auth client for testing
type MockAuthClient struct {
	// Function fields for mocking responses
	IntrospectTokenFunc           func(ctx context.Context, token string) (*TokenIntrospection, error)
	RefreshTokenFunc              func(ctx context.Context, refreshToken string) (*TokenResponse, error)
	RevokeTokenFunc               func(ctx context.Context, token string) error
	ExchangeAuthorizationCodeFunc func(ctx context.Context, code string) (*TokenResponse, error)
	GetAuthorizationURLFunc       func(state string, scopes ...string) string
}

// IntrospectToken implements the auth client interface for mocks
func (m *MockAuthClient) IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error) {
	if m.IntrospectTokenFunc != nil {
		return m.IntrospectTokenFunc(ctx, token)
	}
	return &TokenIntrospection{
		Active:   true,
		Subject:  "test-subject",
		Username: "test-user",
		Email:    "test@example.com",
	}, nil
}

// RefreshToken implements the auth client interface for mocks
func (m *MockAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, refreshToken)
	}
	return &TokenResponse{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresIn:    3600,
		Scope:        "openid profile email",
		ExpiryTime:   time.Now().Add(1 * time.Hour),
	}, nil
}

// RevokeToken implements the auth client interface for mocks
func (m *MockAuthClient) RevokeToken(ctx context.Context, token string) error {
	if m.RevokeTokenFunc != nil {
		return m.RevokeTokenFunc(ctx, token)
	}
	return nil
}

// ExchangeAuthorizationCode implements the auth client interface for mocks
func (m *MockAuthClient) ExchangeAuthorizationCode(ctx context.Context, code string) (*TokenResponse, error) {
	if m.ExchangeAuthorizationCodeFunc != nil {
		return m.ExchangeAuthorizationCodeFunc(ctx, code)
	}
	return &TokenResponse{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresIn:    3600,
		Scope:        "openid profile email",
		ExpiryTime:   time.Now().Add(1 * time.Hour),
	}, nil
}

// GetAuthorizationURL implements the auth client interface for mocks
func (m *MockAuthClient) GetAuthorizationURL(state string, scopes ...string) string {
	if m.GetAuthorizationURLFunc != nil {
		return m.GetAuthorizationURLFunc(state, scopes...)
	}
	return "https://auth.globus.org/v2/oauth2/authorize?mock=true"
}
