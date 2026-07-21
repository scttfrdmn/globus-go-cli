// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package globusauth is the CLI's authentication foundation, built on the v4
// SDK's GlobusApp (app.UserApp) with per-profile, per-resource-server token
// storage. It replaces the previous single-combined-access-token model so that
// each service is authorized with a token minted for its own resource server —
// matching the Python Globus CLI's behavior.
package globusauth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/app"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

// DefaultClientID is the native (public) client used when the user has not
// configured their own. Matches pkg/config.DefaultClientID.
const DefaultClientID = "e6c75d97-532a-4c88-b031-f5a3014430e3"

// Service identifies a Globus service for scope/resource-server lookup.
type Service string

const (
	ServiceAuth     Service = "auth"
	ServiceTransfer Service = "transfer"
	ServiceGroups   Service = "groups"
	ServiceSearch   Service = "search"
	ServiceFlows    Service = "flows"
	ServiceCompute  Service = "compute"
	ServiceTimers   Service = "timers"
)

// serviceInfo holds the scope and resource-server identifier for a service.
type serviceInfo struct {
	scope          string
	resourceServer string
}

// registry maps each service to the scope requested at login and the
// resource-server key its tokens are stored/retrieved under.
var registry = map[Service]serviceInfo{
	ServiceAuth: {
		scope:          "openid profile email",
		resourceServer: "auth.globus.org",
	},
	ServiceTransfer: {
		scope:          "urn:globus:auth:scope:transfer.api.globus.org:all",
		resourceServer: "transfer.api.globus.org",
	},
	ServiceGroups: {
		scope:          "urn:globus:auth:scope:groups.api.globus.org:all",
		resourceServer: "groups.api.globus.org",
	},
	ServiceSearch: {
		scope:          "urn:globus:auth:scope:search.api.globus.org:all",
		resourceServer: "search.api.globus.org",
	},
	ServiceFlows: {
		scope:          "https://auth.globus.org/scopes/eec9b274-0c81-4334-bdc2-54e90e689b9a/manage_flows",
		resourceServer: "flows.globus.org",
	},
	ServiceCompute: {
		scope:          "https://auth.globus.org/scopes/facd7ccc-c5f4-42aa-916b-a0e270e2c2a9/all",
		resourceServer: "funcx_service",
	},
	ServiceTimers: {
		scope:          "https://auth.globus.org/scopes/a1a171d5-48fb-4c77-a7ba-b8c628c20fd5/timers.api",
		resourceServer: "524230d7-ea86-4a52-8312-86065a9e0417",
	},
}

// AllServices is the set of services a full login requests scopes for.
var AllServices = []Service{
	ServiceAuth, ServiceTransfer, ServiceGroups, ServiceSearch,
	ServiceFlows, ServiceCompute, ServiceTimers,
}

// ResourceServer returns the resource-server identifier for a service.
func ResourceServer(s Service) (string, bool) {
	info, ok := registry[s]
	return info.resourceServer, ok
}

// Scope returns the login scope for a service.
func Scope(s Service) (string, bool) {
	info, ok := registry[s]
	return info.scope, ok
}

// tokenStoragePath returns the per-profile token file path
// (~/.globus-cli/tokens/<profile>.json).
func tokenStoragePath(profile string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	if profile == "" {
		profile = "default"
	}
	dir := filepath.Join(home, ".globus-cli", "tokens")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create tokens directory: %w", err)
	}
	return filepath.Join(dir, profile+".json"), nil
}

// NewApp builds a UserApp for the given profile with the provided client
// credentials (clientSecret may be empty for native clients), registering scope
// requirements for the requested services. Tokens are persisted per profile as
// JSON, one entry per resource server.
func NewApp(profile, clientID, clientSecret string, services ...Service) (*app.UserApp, error) {
	if clientID == "" {
		clientID = DefaultClientID
	}
	path, err := tokenStoragePath(profile)
	if err != nil {
		return nil, err
	}
	store, err := tokenstorage.NewJSONTokenStorageWithNamespace(path, "globus-cli")
	if err != nil {
		return nil, fmt.Errorf("cannot open token storage: %w", err)
	}
	userApp, err := app.NewUserApp(clientID, clientSecret, &app.AppConfig{
		TokenStorage:         store,
		RequestRefreshTokens: true,
	})
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		services = AllServices
	}
	for _, svc := range services {
		info, ok := registry[svc]
		if !ok {
			return nil, fmt.Errorf("unknown service %q", svc)
		}
		userApp.AddScopeRequirements(info.resourceServer, info.scope)
	}
	return userApp, nil
}

// Authorizer returns an authorizer for a service's resource server from the
// stored tokens of the given profile. Returns an error advising login if no
// token is stored.
func Authorizer(ctx context.Context, profile, clientID, clientSecret string, svc Service) (core.Authorizer, error) {
	userApp, err := NewApp(profile, clientID, clientSecret, svc)
	if err != nil {
		return nil, err
	}
	info := registry[svc]
	authz, err := userApp.GetAuthorizer(ctx, info.resourceServer)
	if err != nil {
		return nil, fmt.Errorf("%w (run 'globus login')", err)
	}
	return authz, nil
}

// ClientConfig builds a v4 SDK *core.Config authorized for the given service
// from the stored tokens of the profile. Pass the result straight to a service
// package's NewClient(ctx, cfg), e.g.:
//
//	cfg, err := globusauth.ClientConfig(ctx, profile, clientID, secret, globusauth.ServiceTransfer)
//	client, err := transfer.NewClient(ctx, cfg)
//
// The returned config carries the auto-refreshing per-resource-server
// authorizer and the service's scope, so every migrated command constructs its
// client the same way.
func ClientConfig(ctx context.Context, profile, clientID, clientSecret string, svc Service) (*core.Config, error) {
	authz, err := Authorizer(ctx, profile, clientID, clientSecret, svc)
	if err != nil {
		return nil, err
	}
	cfg := &core.Config{Authorizer: authz}
	if scope, ok := Scope(svc); ok {
		cfg.Scopes = []string{scope}
	}
	return cfg, nil
}

// Store opens the profile's raw JSON token storage. Callers that need direct
// access to stored tokens (to display, revoke, or delete them) use this rather
// than going through a UserApp/authorizer.
func Store(profile string) (*tokenstorage.JSONTokenStorage, error) {
	path, err := tokenStoragePath(profile)
	if err != nil {
		return nil, err
	}
	store, err := tokenstorage.NewJSONTokenStorageWithNamespace(path, "globus-cli")
	if err != nil {
		return nil, fmt.Errorf("cannot open token storage: %w", err)
	}
	return store, nil
}

// TokenFor returns the stored token data for a service's resource server, or an
// error advising login if none is stored.
func TokenFor(profile string, svc Service) (*tokenstorage.TokenData, error) {
	info, ok := registry[svc]
	if !ok {
		return nil, fmt.Errorf("unknown service %q", svc)
	}
	store, err := Store(profile)
	if err != nil {
		return nil, err
	}
	data, err := store.Get(info.resourceServer)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("no stored token for %s (run 'globus login')", svc)
	}
	return data, nil
}

// AllTokens returns every stored token (one per resource server) for a profile.
func AllTokens(profile string) ([]*tokenstorage.TokenData, error) {
	store, err := Store(profile)
	if err != nil {
		return nil, err
	}
	return store.GetAll()
}

// RemoveAllTokens deletes every stored token for a profile (used by logout).
func RemoveAllTokens(profile string) error {
	store, err := Store(profile)
	if err != nil {
		return err
	}
	all, err := store.GetAll()
	if err != nil {
		return err
	}
	for _, td := range all {
		if rmErr := store.Remove(td.ResourceServer); rmErr != nil {
			return rmErr
		}
	}
	return nil
}

// AuthClientConfig builds a *core.Config for the Auth service, minted from the
// profile's stored auth token. Convenience wrapper over ClientConfig with
// ServiceAuth.
func AuthClientConfig(ctx context.Context, profile, clientID, clientSecret string) (*core.Config, error) {
	return ClientConfig(ctx, profile, clientID, clientSecret, ServiceAuth)
}

// StoredToken is a minimal view of an OAuth2 token response for one resource
// server, used by StoreTokens to persist tokens obtained outside the UserApp
// login flow (e.g. the device-code flow).
type StoredToken struct {
	ResourceServer string
	AccessToken    string
	RefreshToken   string
	Scope          string
	ExpiresIn      int
	TokenType      string
}

// StoreTokens persists a set of per-resource-server tokens into the profile's
// store, so subsequent commands (which read the store via ClientConfig) find
// them. ExpiresIn is converted to an absolute expiry using the provided now.
func StoreTokens(profile string, now time.Time, tokens ...StoredToken) error {
	store, err := Store(profile)
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if t.ResourceServer == "" || t.AccessToken == "" {
			continue
		}
		if err := store.Store(&tokenstorage.TokenData{
			ResourceServer: t.ResourceServer,
			Scope:          t.Scope,
			AccessToken:    t.AccessToken,
			RefreshToken:   t.RefreshToken,
			ExpiresAt:      now.Add(time.Duration(t.ExpiresIn) * time.Second),
			TokenType:      t.TokenType,
		}); err != nil {
			return err
		}
	}
	return nil
}
