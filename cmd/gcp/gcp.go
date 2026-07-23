// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

// GCPCmd returns the `gcp` command group.
func GCPCmd() *cobra.Command {
	gcpCmd := &cobra.Command{
		Use:   "gcp",
		Short: "Manage Globus Connect Personal endpoints",
		Long: `Manage Globus Connect Personal (GCP) endpoints and collections.

These commands operate on the Globus service (registering and updating GCP
endpoints and their collections); they do NOT install, start, or stop a local
Globus Connect Personal agent. 'gcp create mapped' registers an endpoint with
Globus and prints a setup key you use to configure an installed GCP agent.`,
	}

	gcpCmd.AddCommand(gcpCreateCmd(), gcpSetSubscriptionIDCmd())
	return gcpCmd
}

// --- shared endpoint/collection document flags (mirrors the Python CLI) ---

type gcpDocFlags struct {
	description      string
	organization    string
	department       string
	contactEmail     string
	contactInfo      string
	infoLink         string
	defaultDirectory string
	userMessage      string
	userMessageLink  string
	keywords         []string
	verify           string // force|disable|default
	public           bool
	private          bool
	forceEncryption  bool
	noForceEncrypt   bool
	subscriptionID   string
}

// registerDocFlags adds the endpoint/collection metadata flags shared by the
// GCP create subcommands, matching the Python globus CLI.
func registerDocFlags(cmd *cobra.Command, f *gcpDocFlags, mapped bool) {
	cmd.Flags().StringVar(&f.description, "description", "", "Description for the collection")
	cmd.Flags().StringVar(&f.organization, "organization", "", "Organization for the collection")
	cmd.Flags().StringVar(&f.department, "department", "", "Department which operates the collection")
	cmd.Flags().StringVar(&f.contactEmail, "contact-email", "", "Contact email for the collection")
	cmd.Flags().StringVar(&f.contactInfo, "contact-info", "", "Contact info for the collection")
	cmd.Flags().StringVar(&f.infoLink, "info-link", "", "Link for info about the collection")
	cmd.Flags().StringVar(&f.defaultDirectory, "default-directory", "", "Default directory when browsing the collection")
	cmd.Flags().StringVar(&f.userMessage, "user-message", "", "A message for clients to display to users")
	cmd.Flags().StringVar(&f.userMessageLink, "user-message-link", "", "Link to additional messaging for clients to display to users")
	cmd.Flags().StringSliceVar(&f.keywords, "keywords", nil, "Comma-separated keywords to help searches find the collection")
	cmd.Flags().StringVar(&f.verify, "verify", "", "Set checksum verification: force, disable, or default")
	cmd.Flags().BoolVar(&f.forceEncryption, "force-encryption", false, "Require encryption for all transfers on the collection")
	cmd.Flags().BoolVar(&f.noForceEncrypt, "no-force-encryption", false, "Do not require encryption for transfers")
	if mapped {
		cmd.Flags().BoolVar(&f.public, "public", false, "Set the collection to be public")
		cmd.Flags().BoolVar(&f.private, "private", false, "Set the collection to be private")
		cmd.Flags().StringVar(&f.subscriptionID, "subscription-id", "", "Set the collection as managed with the given subscription ID")
	}
}

// applyDocFlags sets the JSON keys on an endpoint document for the flags the
// user actually provided.
func applyDocFlags(cmd *cobra.Command, f *gcpDocFlags, doc map[string]interface{}) error {
	set := func(name, key, val string) {
		if cmd.Flags().Changed(name) {
			doc[key] = val
		}
	}
	set("description", "description", f.description)
	set("organization", "organization", f.organization)
	set("department", "department", f.department)
	set("contact-email", "contact_email", f.contactEmail)
	set("contact-info", "contact_info", f.contactInfo)
	set("info-link", "info_link", f.infoLink)
	set("default-directory", "default_directory", f.defaultDirectory)
	set("user-message", "user_message", f.userMessage)
	set("user-message-link", "user_message_link", f.userMessageLink)
	if cmd.Flags().Changed("keywords") {
		doc["keywords"] = strings.Join(f.keywords, ",")
	}
	switch f.verify {
	case "":
	case "force":
		doc["force_verify"] = true
		doc["disable_verify"] = false
	case "disable":
		doc["disable_verify"] = true
		doc["force_verify"] = false
	case "default":
		doc["force_verify"] = false
		doc["disable_verify"] = false
	default:
		return fmt.Errorf("invalid --verify value %q (use force, disable, or default)", f.verify)
	}
	if cmd.Flags().Changed("force-encryption") {
		doc["force_encryption"] = f.forceEncryption
	}
	if cmd.Flags().Changed("no-force-encryption") && f.noForceEncrypt {
		doc["force_encryption"] = false
	}
	if f.public {
		doc["public"] = true
	}
	if f.private {
		doc["public"] = false
	}
	if cmd.Flags().Changed("subscription-id") {
		doc["subscription_id"] = f.subscriptionID
	}
	return nil
}

// --- gcp create (mapped | guest) ---

func gcpCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Globus Connect Personal collection",
	}
	createCmd.AddCommand(gcpCreateMappedCmd(), gcpCreateGuestCmd())
	return createCmd
}

func gcpCreateMappedCmd() *cobra.Command {
	var f gcpDocFlags
	cmd := &cobra.Command{
		Use:   "mapped DISPLAY_NAME",
		Short: "Register a new Globus Connect Personal mapped collection (endpoint)",
		Long: `Register a new Globus Connect Personal mapped collection with Globus.

In GCP, the mapped collection and the endpoint are the same object. This does
not install or start a local GCP agent — it performs the registration and
prints a setup key you use to configure an installed agent.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":         "endpoint",
				"display_name":      args[0],
				"is_globus_connect": true,
			}
			if err := applyDocFlags(cmd, &f, doc); err != nil {
				return err
			}

			resp, err := client.CreateEndpoint(ctx, doc)
			if err != nil {
				return fmt.Errorf("failed to create GCP endpoint: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
				return formatter.FormatOutput(resp, nil)
			}
			if id, ok := resp["id"].(string); ok {
				fmt.Fprintf(cmd.OutOrStdout(), "Endpoint ID: %s\n", id)
			}
			if key, ok := resp["globus_connect_setup_key"].(string); ok {
				fmt.Fprintf(cmd.OutOrStdout(), "Setup Key:   %s\n", key)
			}
			return nil
		},
	}
	registerDocFlags(cmd, &f, true)
	return cmd
}

func gcpCreateGuestCmd() *cobra.Command {
	var f gcpDocFlags
	cmd := &cobra.Command{
		Use:   "guest DISPLAY_NAME HOST_ENDPOINT_ID:PATH",
		Short: "Create a guest collection on a Globus Connect Personal endpoint",
		Long: `Create a guest collection rooted at a path on a Globus Connect Personal
host endpoint.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			hostID, hostPath, ok := strings.Cut(args[1], ":")
			if !ok || hostID == "" || hostPath == "" {
				return fmt.Errorf("second argument must be HOST_ENDPOINT_ID:PATH")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":    "shared_endpoint",
				"display_name": args[0],
				"host_endpoint": hostID,
				"host_path":    hostPath,
			}
			if err := applyDocFlags(cmd, &f, doc); err != nil {
				return err
			}

			resp, err := client.CreateSharedEndpoint(ctx, doc)
			if err != nil {
				return fmt.Errorf("failed to create guest collection: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
				return formatter.FormatOutput(resp, nil)
			}
			if id, ok := resp["id"].(string); ok {
				fmt.Fprintf(cmd.OutOrStdout(), "Collection ID: %s\n", id)
			}
			return nil
		},
	}
	registerDocFlags(cmd, &f, false)
	return cmd
}

// --- gcp set-subscription-id ---

func gcpSetSubscriptionIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-subscription-id ENDPOINT_ID SUBSCRIPTION_ID",
		Short: "Set the subscription for a Globus Connect Personal endpoint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			if _, err := client.SetSubscriptionID(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("failed to set subscription: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Set subscription %s on endpoint %s\n", args[1], args[0])
			return nil
		},
	}
}
