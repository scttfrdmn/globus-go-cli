// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

// endpointUpdateCmd returns the "endpoint update" command, which patches an
// endpoint document with only the fields the user explicitly set.
func endpointUpdateCmd() *cobra.Command {
	var (
		displayName          string
		description          string
		organization         string
		department           string
		contactEmail         string
		contactInfo          string
		infoLink             string
		defaultDirectory     string
		noDefaultDirectory   bool
		keywords             []string
		public               bool
		private              bool
		forceEncryption      bool
		noForceEncryption    bool
		disableVerify        bool
		noDisableVerify      bool
		subscriptionID       string
		noManaged            bool
		managed              bool
		networkUse           string
		maxConcurrency       int
		preferredConcurrency int
		maxParallelism       int
		preferredParallelism int
		userMessage          string
		userMessageLink      string
		oauthServer          string
		myproxyServer        string
		myproxyDN            string
		location             string
	)

	cmd := &cobra.Command{
		Use:   "update ENDPOINT_ID",
		Short: "Update a Globus endpoint",
		Long: `Update mutable fields on a Globus endpoint.

Only the flags you provide are sent; unset fields are left unchanged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{"DATA_TYPE": "endpoint"}
			if cmd.Flags().Changed("display-name") {
				doc["display_name"] = displayName
			}
			if cmd.Flags().Changed("description") {
				doc["description"] = description
			}
			if cmd.Flags().Changed("organization") {
				doc["organization"] = organization
			}
			if cmd.Flags().Changed("department") {
				doc["department"] = department
			}
			if cmd.Flags().Changed("contact-email") {
				doc["contact_email"] = contactEmail
			}
			if cmd.Flags().Changed("contact-info") {
				doc["contact_info"] = contactInfo
			}
			if cmd.Flags().Changed("info-link") {
				doc["info_link"] = infoLink
			}
			if cmd.Flags().Changed("default-directory") {
				doc["default_directory"] = defaultDirectory
			}
			if noDefaultDirectory {
				doc["default_directory"] = nil
			}
			if cmd.Flags().Changed("keywords") {
				doc["keywords"] = keywords
			}
			// --public / --private
			if cmd.Flags().Changed("public") {
				doc["public"] = public
			}
			if private {
				doc["public"] = false
			}
			// --force-encryption / --no-force-encryption
			if cmd.Flags().Changed("force-encryption") {
				doc["force_encryption"] = forceEncryption
			}
			if noForceEncryption {
				doc["force_encryption"] = false
			}
			// --disable-verify / --no-disable-verify
			if cmd.Flags().Changed("disable-verify") {
				doc["disable_verify"] = disableVerify
			}
			if noDisableVerify {
				doc["disable_verify"] = false
			}
			// Managed-endpoint / subscription handling.
			if cmd.Flags().Changed("subscription-id") {
				doc["subscription_id"] = subscriptionID
			}
			if noManaged {
				doc["subscription_id"] = nil
			}
			if managed {
				return fmt.Errorf("--managed is not supported: it requires resolving your subscription ID; use --subscription-id instead")
			}
			if cmd.Flags().Changed("network-use") {
				doc["network_use"] = networkUse
			}
			if cmd.Flags().Changed("max-concurrency") {
				doc["max_concurrency"] = maxConcurrency
			}
			if cmd.Flags().Changed("preferred-concurrency") {
				doc["preferred_concurrency"] = preferredConcurrency
			}
			if cmd.Flags().Changed("max-parallelism") {
				doc["max_parallelism"] = maxParallelism
			}
			if cmd.Flags().Changed("preferred-parallelism") {
				doc["preferred_parallelism"] = preferredParallelism
			}
			if cmd.Flags().Changed("user-message") {
				doc["user_message"] = userMessage
			}
			if cmd.Flags().Changed("user-message-link") {
				doc["user_message_link"] = userMessageLink
			}
			if cmd.Flags().Changed("oauth-server") {
				doc["oauth_server"] = oauthServer
			}
			if cmd.Flags().Changed("myproxy-server") {
				doc["myproxy_server"] = myproxyServer
			}
			if cmd.Flags().Changed("myproxy-dn") {
				doc["myproxy_dn"] = myproxyDN
			}
			if cmd.Flags().Changed("location") {
				doc["location"] = location
			}

			resp, err := client.UpdateEndpoint(ctx, args[0], doc)
			if err != nil {
				return fmt.Errorf("failed to update endpoint: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Endpoint %s updated.\n", args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "Name for the endpoint")
	cmd.Flags().StringVar(&description, "description", "", "Description for the endpoint")
	cmd.Flags().StringVar(&organization, "organization", "", "Organization for the endpoint")
	cmd.Flags().StringVar(&department, "department", "", "Department which operates the endpoint")
	cmd.Flags().StringVar(&contactEmail, "contact-email", "", "Contact email for the endpoint")
	cmd.Flags().StringVar(&contactInfo, "contact-info", "", "Contact info for the endpoint")
	cmd.Flags().StringVar(&infoLink, "info-link", "", "Link for info about the endpoint")
	cmd.Flags().StringVar(&defaultDirectory, "default-directory", "", "Default directory when browsing or executing tasks on the endpoint")
	cmd.Flags().BoolVar(&noDefaultDirectory, "no-default-directory", false, "Unset any default directory on the endpoint")
	cmd.Flags().StringSliceVar(&keywords, "keywords", nil, "Comma separated list of keywords to help searches for the endpoint")
	cmd.Flags().BoolVar(&public, "public", false, "Set the endpoint to be public")
	cmd.Flags().BoolVar(&private, "private", false, "Set the endpoint to be private")
	cmd.Flags().BoolVar(&forceEncryption, "force-encryption", false, "Force the endpoint to encrypt transfers")
	cmd.Flags().BoolVar(&noForceEncryption, "no-force-encryption", false, "Do not force the endpoint to encrypt transfers")
	cmd.Flags().BoolVar(&disableVerify, "disable-verify", false, "Set the endpoint to ignore checksum verification")
	cmd.Flags().BoolVar(&noDisableVerify, "no-disable-verify", false, "Do not ignore checksum verification")
	cmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Set the endpoint as managed with the given subscription ID")
	cmd.Flags().BoolVar(&noManaged, "no-managed", false, "Unset the endpoint as a managed endpoint")
	cmd.Flags().BoolVar(&managed, "managed", false, "Set the endpoint as a managed endpoint (requires --subscription-id)")
	cmd.Flags().StringVar(&networkUse, "network-use", "", "Network use level (normal, minimal, aggressive, custom)")
	cmd.Flags().IntVar(&maxConcurrency, "max-concurrency", 0, "Endpoint max concurrency; requires --network-use=custom")
	cmd.Flags().IntVar(&preferredConcurrency, "preferred-concurrency", 0, "Endpoint preferred concurrency; requires --network-use=custom")
	cmd.Flags().IntVar(&maxParallelism, "max-parallelism", 0, "Endpoint max parallelism; requires --network-use=custom")
	cmd.Flags().IntVar(&preferredParallelism, "preferred-parallelism", 0, "Endpoint preferred parallelism; requires --network-use=custom")
	cmd.Flags().StringVar(&userMessage, "user-message", "", "A message for clients to display to users when interacting with this endpoint")
	cmd.Flags().StringVar(&userMessageLink, "user-message-link", "", "Link to additional messaging for clients to display to users")
	cmd.Flags().StringVar(&oauthServer, "oauth-server", "", "Set the OAuth Server URI (Globus Connect Server only)")
	cmd.Flags().StringVar(&myproxyServer, "myproxy-server", "", "Set the MyProxy Server URI (Globus Connect Server only)")
	cmd.Flags().StringVar(&myproxyDN, "myproxy-dn", "", "Set the MyProxy Server DN (Globus Connect Server only)")
	cmd.Flags().StringVar(&location, "location", "", "Manually set the endpoint's LATITUDE,LONGITUDE (Globus Connect Server only)")

	return cmd
}

// endpointDeleteCmd returns the "endpoint delete" command.
func endpointDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID",
		Short: "Delete a Globus endpoint",
		Long:  `Delete a Globus endpoint by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.DeleteEndpoint(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to delete endpoint: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Endpoint %s deleted.\n", args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// endpointRoleCmd returns the "endpoint role" command group for managing
// endpoint role assignments.
func endpointRoleCmd() *cobra.Command {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage endpoint role assignments",
		Long:  `List, show, create, and delete role assignments on a Globus endpoint.`,
	}

	roleCmd.AddCommand(
		endpointRoleListCmd(),
		endpointRoleShowCmd(),
		endpointRoleCreateCmd(),
		endpointRoleDeleteCmd(),
	)

	return roleCmd
}

func endpointRoleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List role assignments for an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointRoleList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list endpoint roles: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointRoleShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID ROLE_ID",
		Short: "Show a role assignment for an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.GetEndpointRole(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to get endpoint role: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointRoleCreateCmd() *cobra.Command {
	var (
		principal     string
		principalType string
		role          string
		identity      string
		group         string
	)

	cmd := &cobra.Command{
		Use:   "create ENDPOINT_ID",
		Short: "Create a role assignment on an endpoint",
		Long: `Assign a role to a principal on a Globus endpoint.

The --role value is one of administrator, access_manager, activity_manager,
or activity_monitor.

Specify the security principal with exactly one of --identity, --group, or the
lower-level --principal/--principal-type pair.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			// Resolve the principal from --identity/--group (Python-compatible)
			// or the explicit --principal/--principal-type pair.
			if identity != "" {
				principal = identity
				principalType = "identity"
			} else if group != "" {
				principal = group
				principalType = "group"
			}
			if principal == "" {
				return fmt.Errorf("one of --identity, --group, or --principal is required")
			}

			doc := map[string]interface{}{
				"DATA_TYPE":      "role",
				"principal_type": principalType,
				"principal":      principal,
				"role":           role,
			}

			resp, err := client.AddEndpointRole(ctx, args[0], doc)
			if err != nil {
				return fmt.Errorf("failed to create endpoint role: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&identity, "identity", "", "Identity ID to use as the security principal")
	cmd.Flags().StringVar(&group, "group", "", "Group ID to use as the security principal")
	cmd.Flags().StringVar(&principal, "principal", "", "Principal (identity or group ID) to assign the role to")
	cmd.Flags().StringVar(&principalType, "principal-type", "identity", "Principal type (identity or group)")
	cmd.Flags().StringVar(&role, "role", "", "Role name (administrator, access_manager, activity_manager, activity_monitor)")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func endpointRoleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID ROLE_ID",
		Short: "Delete a role assignment from an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.DeleteEndpointRole(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to delete endpoint role: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Role %s deleted from endpoint %s.\n", args[1], args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// endpointPermissionCmd returns the "endpoint permission" command group for
// managing endpoint access rules (ACLs).
func endpointPermissionCmd() *cobra.Command {
	permCmd := &cobra.Command{
		Use:   "permission",
		Short: "Manage endpoint access rules (ACLs)",
		Long:  `List, show, create, update, and delete access rules on a Globus endpoint.`,
	}

	permCmd.AddCommand(
		endpointPermissionListCmd(),
		endpointPermissionShowCmd(),
		endpointPermissionCreateCmd(),
		endpointPermissionUpdateCmd(),
		endpointPermissionDeleteCmd(),
	)

	return permCmd
}

func endpointPermissionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list ENDPOINT_ID",
		Short: "List access rules for an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointACLList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list endpoint access rules: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointPermissionShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID RULE_ID",
		Short: "Show an access rule for an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.GetEndpointACLRule(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to get endpoint access rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointPermissionCreateCmd() *cobra.Command {
	var (
		permissions      string
		principal        string
		principalType    string
		path             string
		identity         string
		group            string
		allAuthenticated bool
		anonymous        bool
		notifyEmail      string
		notifyMessage    string
		expirationDate   string
	)

	cmd := &cobra.Command{
		Use:   "create ENDPOINT_ID",
		Short: "Create an access rule on an endpoint",
		Long: `Create an access rule (ACL) granting permissions on a path.

Specify the security principal with exactly one of --identity, --group,
--all-authenticated, --anonymous, or the lower-level --principal/--principal-type
pair. The --principal-type value is one of identity, group,
all_authenticated_users, or anonymous.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			// Resolve the principal from the Python-compatible convenience flags
			// or the explicit --principal/--principal-type pair.
			switch {
			case identity != "":
				principal = identity
				principalType = "identity"
			case group != "":
				principal = group
				principalType = "group"
			case allAuthenticated:
				principal = ""
				principalType = "all_authenticated_users"
			case anonymous:
				principal = ""
				principalType = "anonymous"
			}

			doc := map[string]interface{}{
				"DATA_TYPE":      "access",
				"principal_type": principalType,
				"principal":      principal,
				"path":           path,
				"permissions":    permissions,
			}
			if cmd.Flags().Changed("notify-email") {
				doc["notify_email"] = notifyEmail
			}
			if cmd.Flags().Changed("notify-message") {
				doc["notify_message"] = notifyMessage
			}
			if cmd.Flags().Changed("expiration-date") {
				doc["expiration_date"] = expirationDate
			}

			resp, err := client.AddEndpointACLRule(ctx, args[0], doc)
			if err != nil {
				return fmt.Errorf("failed to create endpoint access rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&permissions, "permissions", "", `Permissions to add: "r" (Read-Only) or "rw" (Read/Write)`)
	cmd.Flags().StringVar(&identity, "identity", "", "Identity ID to use as the security principal")
	cmd.Flags().StringVar(&group, "group", "", "Group ID to use as the security principal")
	cmd.Flags().BoolVar(&allAuthenticated, "all-authenticated", false, "Allow anyone access, as long as they log in")
	cmd.Flags().BoolVar(&anonymous, "anonymous", false, "Allow anyone access, even without logging in")
	cmd.Flags().StringVar(&principal, "principal", "", "Principal (identity or group ID) to grant access to")
	cmd.Flags().StringVar(&principalType, "principal-type", "identity", "Principal type (identity, group, all_authenticated_users, anonymous)")
	cmd.Flags().StringVar(&path, "path", "", "Path the rule applies to")
	cmd.Flags().StringVar(&notifyEmail, "notify-email", "", "An email address to notify that the permission has been added")
	cmd.Flags().StringVar(&notifyMessage, "notify-message", "", "A custom message to add to email notifications")
	cmd.Flags().StringVar(&expirationDate, "expiration-date", "", "Expiration date for the permission in ISO 8601 format")
	_ = cmd.MarkFlagRequired("permissions")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func endpointPermissionUpdateCmd() *cobra.Command {
	var (
		permissions    string
		expirationDate string
	)

	cmd := &cobra.Command{
		Use:   "update ENDPOINT_ID RULE_ID",
		Short: "Update an access rule on an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":   "access",
				"permissions": permissions,
			}
			if cmd.Flags().Changed("expiration-date") {
				doc["expiration_date"] = expirationDate
			}

			resp, err := client.UpdateEndpointACLRule(ctx, args[0], args[1], doc)
			if err != nil {
				return fmt.Errorf("failed to update endpoint access rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&permissions, "permissions", "", `Permissions to add: "r" (Read-Only) or "rw" (Read/Write)`)
	cmd.Flags().StringVar(&expirationDate, "expiration-date", "", "Expiration date for the permission in ISO 8601 format")
	_ = cmd.MarkFlagRequired("permissions")

	return cmd
}

func endpointPermissionDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT_ID RULE_ID",
		Short: "Delete an access rule from an endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.DeleteEndpointACLRule(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to delete endpoint access rule: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Access rule %s deleted from endpoint %s.\n", args[1], args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// endpointSetSubscriptionIDCmd returns the "endpoint set-subscription-id"
// command, which associates an endpoint with a managed endpoint subscription.
func endpointSetSubscriptionIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-subscription-id ENDPOINT_ID SUBSCRIPTION_ID",
		Short: "Set the subscription ID for an endpoint",
		Long:  `Associate a Globus endpoint with a managed endpoint subscription.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.SetSubscriptionID(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to set subscription ID: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Endpoint %s subscription ID set to %s.\n", args[0], args[1])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// endpointMySharedEndpointListCmd returns the "endpoint my-shared-endpoint-list"
// command, which lists shared endpoints the current user has created on a host.
func endpointMySharedEndpointListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "my-shared-endpoint-list ENDPOINT_ID",
		Short: "List shared endpoints you created on a host endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.MySharedEndpointList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list shared endpoints: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

// formatGenericResponse routes a GenericResponse (map[string]interface{})
// through the shared formatter so -F (text/json/unix) and --jmespath/--jq
// work uniformly. For text/unix, the whole map is emitted.
func formatGenericResponse(cmd *cobra.Command, resp map[string]interface{}) error {
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.FormatOutput(resp, nil)
}

// printResponseCodeMessage prints the "code" and "message" fields of a Transfer
// GenericResponse when present, for success/status feedback on mutations.
func printResponseCodeMessage(cmd *cobra.Command, resp map[string]interface{}) {
	if resp == nil {
		return
	}
	if code, ok := resp["code"].(string); ok && code != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Code:    %s\n", code)
	}
	if msg, ok := resp["message"].(string); ok && msg != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Message: %s\n", msg)
	}
}
