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
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/login"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

// SessionParams carries the Globus Auth session-enforcement (step-up auth)
// options for SessionLogin — the parameters behind a `session update`.
type SessionParams struct {
	RequiredIdentities   []string
	RequiredSingleDomain []string
	RequiredPolicies     []string
	RequiredMFA          bool
	Message              string
}

// SessionLogin runs the command-line OAuth2 login flow for the given scopes and
// (optional) session-enforcement parameters, then stores the resulting tokens
// in the profile's store. It powers `session update` (force step-up re-auth via
// the session params) and `session consent` (grant a specific scope). Scopes
// must be non-empty. Returns the resource servers for which tokens were stored.
func SessionLogin(ctx context.Context, profile, clientID, clientSecret string, scopes []string, sp *SessionParams) ([]string, error) {
	if len(scopes) == 0 {
		return nil, fmt.Errorf("at least one scope is required")
	}
	if clientID == "" {
		clientID = DefaultClientID
	}

	mgr := login.NewCommandLineLoginFlowManager(clientID, clientSecret)
	params := login.AuthParams{
		Scopes:         scopes,
		RequestRefresh: true,
	}
	if sp != nil {
		params.SessionRequiredIdentities = sp.RequiredIdentities
		params.SessionRequiredSingleDomain = sp.RequiredSingleDomain
		params.SessionRequiredPolicies = sp.RequiredPolicies
		params.SessionRequiredMFA = sp.RequiredMFA
		params.SessionMessage = sp.Message
	}

	result, err := mgr.RunLoginFlow(ctx, params)
	if err != nil {
		return nil, err
	}

	store, err := Store(profile)
	if err != nil {
		return nil, err
	}
	var stored []string
	for _, td := range result.Tokens {
		if err := store.Store(td); err != nil {
			return nil, fmt.Errorf("store token for %s: %w", td.ResourceServer, err)
		}
		stored = append(stored, td.ResourceServer)
	}
	return stored, nil
}

// DefaultClientID is the native (public) client used when the user has not
// configured their own. Matches pkg/config.DefaultClientID.
const DefaultClientID = "ccc07ea1-bfff-4ac0-b36e-da0141ca01c5"

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

// AllServices is every service the CLI knows how to authorize. It is NOT the
// default login set: ServiceTimers uses a client-specific scope
// (.../timers.api) that a generic native/public client (including the CLI's
// default client) is not authorized to request, so requesting it in a full
// login makes Globus reject the whole login with UNKNOWN_SCOPE_ERROR (issue
// #40). Use DefaultLoginServices for the out-of-box login; request Timers
// explicitly via `login --scopes timers` when the configured client supports
// it.
var AllServices = []Service{
	ServiceAuth, ServiceTransfer, ServiceGroups, ServiceSearch,
	ServiceFlows, ServiceCompute, ServiceTimers,
}

// DefaultLoginServices is the set a plain `globus login` requests — AllServices
// minus ServiceTimers (see AllServices for why). Every scope here is one the
// default native client can request, so login succeeds out of the box.
var DefaultLoginServices = []Service{
	ServiceAuth, ServiceTransfer, ServiceGroups, ServiceSearch,
	ServiceFlows, ServiceCompute,
}

// ServiceByName maps a service name (the registry key, e.g. "transfer",
// "timers") to its Service. Used to resolve `login --scopes` entries.
func ServiceByName(name string) (Service, bool) {
	svc := Service(name)
	if _, ok := registry[svc]; ok {
		return svc, true
	}
	return "", false
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
		services = DefaultLoginServices
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

// ScopedApp builds a UserApp registered for a single dynamic (resourceServer,
// scope) pair rather than the fixed service registry. Used for GCS, whose
// scopes are keyed on a specific endpoint or collection ID rather than a
// well-known service resource server.
func ScopedApp(profile, clientID, clientSecret, resourceServer, scope string) (*app.UserApp, error) {
	return ScopedAppWithNamespace(profile, clientID, clientSecret, resourceServer, scope, "globus-cli")
}

// ScopedAppWithNamespace is like ScopedApp but stores/reads the consent token
// under a caller-chosen token-storage namespace. This matters when a dynamic
// scope shares a resource server with the standard login token (e.g. the
// auth.globus.org manage_projects scope vs the login's openid/profile/email
// token): a distinct namespace keeps them from overwriting each other and
// prevents GetAuthorizer from returning the wrong token.
func ScopedAppWithNamespace(profile, clientID, clientSecret, resourceServer, scope, namespace string) (*app.UserApp, error) {
	if clientID == "" {
		clientID = DefaultClientID
	}
	path, err := tokenStoragePath(profile)
	if err != nil {
		return nil, err
	}
	store, err := tokenstorage.NewJSONTokenStorageWithNamespace(path, namespace)
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
	userApp.AddScopeRequirements(resourceServer, scope)
	return userApp, nil
}

// ScopedSessionConsentConfig re-drives a consent for a single (resourceServer,
// scope) pair under a caller-chosen namespace, forcing step-up authentication
// with the given session-enforcement parameters (e.g. session_required_policies
// returned by a 403), then stores the resulting token in that namespace and
// returns a *core.Config authorized with it.
//
// Unlike ScopedClientConfigWithNamespace (which only escalates consent when no
// token is stored), this always runs the login flow — it is meant to be called
// after an operation fails with a session-policy requirement, so the fresh
// token carries the required policies in session. The token is stored in the
// SAME namespace the scoped config reads from, so the retried operation picks
// it up.
func ScopedSessionConsentConfig(ctx context.Context, profile, clientID, clientSecret, resourceServer, scope, namespace string, sp *SessionParams) (*core.Config, error) {
	if clientID == "" {
		clientID = DefaultClientID
	}

	mgr := login.NewCommandLineLoginFlowManager(clientID, clientSecret)
	params := login.AuthParams{
		Scopes:         []string{scope},
		RequestRefresh: true,
	}
	if sp != nil {
		params.SessionRequiredIdentities = sp.RequiredIdentities
		params.SessionRequiredSingleDomain = sp.RequiredSingleDomain
		params.SessionRequiredPolicies = sp.RequiredPolicies
		params.SessionRequiredMFA = sp.RequiredMFA
		params.SessionMessage = sp.Message
	}

	result, err := mgr.RunLoginFlow(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("consent login failed: %w", err)
	}

	path, err := tokenStoragePath(profile)
	if err != nil {
		return nil, err
	}
	store, err := tokenstorage.NewJSONTokenStorageWithNamespace(path, namespace)
	if err != nil {
		return nil, fmt.Errorf("cannot open token storage: %w", err)
	}
	for _, td := range result.Tokens {
		if err := store.Store(td); err != nil {
			return nil, fmt.Errorf("store token for %s: %w", td.ResourceServer, err)
		}
	}

	// Build a config from the freshly stored token for the target resource server.
	userApp, err := ScopedAppWithNamespace(profile, clientID, clientSecret, resourceServer, scope, namespace)
	if err != nil {
		return nil, err
	}
	authz, err := userApp.GetAuthorizer(ctx, resourceServer)
	if err != nil {
		return nil, fmt.Errorf("no token after consent for %q: %w", resourceServer, err)
	}
	return &core.Config{Authorizer: authz, Scopes: []string{scope}}, nil
}

// ScopedClientConfig returns a *core.Config authorized for a dynamic
// (resourceServer, scope) pair from the profile's stored tokens — used to build
// a GCS CollectionClient. If no token is stored for that resource server and
// allowConsent is true, it runs a consent login (the paste-code flow) to obtain
// one; otherwise it returns an error advising the user to run the login. The
// scope determines what the consent grants (an endpoint's manage_collections or
// a collection's data_access).
func ScopedClientConfig(ctx context.Context, profile, clientID, clientSecret, resourceServer, scope string, allowConsent bool) (*core.Config, error) {
	return ScopedClientConfigWithNamespace(ctx, profile, clientID, clientSecret, resourceServer, scope, "globus-cli", allowConsent)
}

// ScopedClientConfigWithNamespace is ScopedClientConfig with an explicit
// token-storage namespace (see ScopedAppWithNamespace for why this is needed
// when a scope shares a resource server with the login token).
func ScopedClientConfigWithNamespace(ctx context.Context, profile, clientID, clientSecret, resourceServer, scope, namespace string, allowConsent bool) (*core.Config, error) {
	userApp, err := ScopedAppWithNamespace(profile, clientID, clientSecret, resourceServer, scope, namespace)
	if err != nil {
		return nil, err
	}

	authz, err := userApp.GetAuthorizer(ctx, resourceServer)
	if err != nil {
		if !allowConsent {
			return nil, fmt.Errorf("%w (run 'globus login' or retry with consent)", err)
		}
		// No stored token for this resource server: escalate consent.
		if lerr := userApp.Login(ctx); lerr != nil {
			return nil, fmt.Errorf("consent login failed: %w\n\n"+
				"If the browser rejected the request with \"Invalid client_id\" or "+
				"\"PKCE code_challenge required\", the configured client cannot complete "+
				"this consent. Set a native/public client registered for the "+
				"https://auth.globus.org/v2/web/auth-code redirect via GLOBUS_CLIENT_ID "+
				"or your profile config.", lerr)
		}
		authz, err = userApp.GetAuthorizer(ctx, resourceServer)
		if err != nil {
			return nil, fmt.Errorf("no token after consent for %q: %w", resourceServer, err)
		}
	}

	return &core.Config{Authorizer: authz, Scopes: []string{scope}}, nil
}
