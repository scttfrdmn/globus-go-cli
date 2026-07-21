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
