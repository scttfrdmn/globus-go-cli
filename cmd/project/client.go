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
	"fmt"

	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/globusauth"
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
