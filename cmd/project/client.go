// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package project implements the `globus project` command tree: management of
// Globus Auth projects, their registered clients (service accounts / app
// registrations), and client secret credentials — the "developer console"
// administrative surface. It ports the standalone globus-project-manager tool
// onto the v4 SDK.
//
// These operations require the auth.globus.org `manage_projects` scope, which
// the standard `globus login` scope set does not include. getClient obtains a
// manage_projects authorizer via a dedicated consent (escalated on first use),
// stored under its own token-storage namespace so it never collides with the
// login token (both live on the auth.globus.org resource server).
package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

const (
	// manageProjectsScope is the Globus Auth scope required for project /
	// client / credential management.
	manageProjectsScope = "urn:globus:auth:scope:auth.globus.org:manage_projects"
	// authResourceServer is the resource server manage_projects tokens key on.
	authResourceServer = "auth.globus.org"
	// manageProjectsNamespace isolates the manage_projects consent token from
	// the login token (both are on auth.globus.org).
	manageProjectsNamespace = "globus-cli-manage-projects"
)

// getClient builds a v4 Auth client authorized with the manage_projects scope
// for the current profile, escalating consent (a paste-code login) on first
// use. Used by every project/client/credential command.
func getClient(ctx context.Context) (*auth.Client, error) {
	profile := viper.GetString("profile")

	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	cfg, err := globusauth.ScopedClientConfigWithNamespace(
		ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret,
		authResourceServer, manageProjectsScope, manageProjectsNamespace, true,
	)
	if err != nil {
		return nil, fmt.Errorf("could not obtain manage_projects consent: %w", err)
	}

	client, err := auth.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}
	return client, nil
}

// getClientWithPolicies builds a manage_projects Auth client whose token was
// obtained under a consent that satisfies the given session_required_policies
// (high-assurance / session freshness). It re-drives the manage_projects
// consent in the isolated namespace with those policies, so the resulting token
// carries them in session. Used by withProjectRetry after a 403.
func getClientWithPolicies(ctx context.Context, policies []string) (*auth.Client, error) {
	profile := viper.GetString("profile")

	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	cfg, err := globusauth.ScopedSessionConsentConfig(
		ctx, profile, clientCfg.ClientID, clientCfg.ClientSecret,
		authResourceServer, manageProjectsScope, manageProjectsNamespace,
		&globusauth.SessionParams{RequiredPolicies: policies},
	)
	if err != nil {
		return nil, fmt.Errorf("could not obtain policy-satisfying manage_projects consent: %w", err)
	}

	client, err := auth.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}
	return client, nil
}

// withProjectRetry runs a project operation and, if it fails with a 403 that
// carries session_required_policies (a high-assurance auth-policy requirement),
// re-drives the manage_projects consent with those policies and retries once.
// This mirrors the Python globus-cli, which auto-reauthenticates against the
// returned authorization_parameters. Callers pass a closure that takes a client
// and performs the operation; op is invoked at most twice.
func withProjectRetry(ctx context.Context, op func(client *auth.Client) error) error {
	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	err = op(client)
	if err == nil {
		return nil
	}
	policies := sessionRequiredPolicies(err)
	if len(policies) == 0 {
		return err
	}
	fmt.Println("This project requires a high-assurance session; re-authenticating to satisfy its policy...")
	client, cerr := getClientWithPolicies(ctx, policies)
	if cerr != nil {
		// Surface the original 403 alongside the re-auth failure.
		return fmt.Errorf("%w (re-auth for required policy failed: %v)", err, cerr)
	}
	return op(client)
}

// sessionRequiredPolicies extracts authorization_parameters.session_required_policies
// from a Globus API 403 error, if present. Globus returns these on the error
// body when an auth policy demands a fresh, policy-scoped session. Returns nil
// when the error is not such a 403.
func sessionRequiredPolicies(err error) []string {
	var apiErr *core.APIError
	if !errors.As(err, &apiErr) {
		return nil
	}
	if apiErr.StatusCode != 403 || apiErr.Details == nil {
		return nil
	}
	ap, ok := apiErr.Details["authorization_parameters"].(map[string]interface{})
	if !ok {
		return nil
	}
	return toStringSlice(ap["session_required_policies"])
}

// toStringSlice coerces a decoded JSON value that may be a []interface{} of
// strings, or a single comma-joined string, into a []string.
func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	default:
		return nil
	}
}
