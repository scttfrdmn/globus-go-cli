// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
)

// Choices accepted by "flows registered-api list", matching Python globus-cli 3.42.0.
var (
	registeredAPIRoles         = []string{"owner", "administrator", "viewer"}
	registeredAPIOrderByFields = []string{"id", "name", "created_timestamp", "updated_timestamp"}
)

// registeredAPIListPage is the JSON document emitted by "registered-api list
// -F json", matching the Python CLI's json_conversion_key ("registered_apis").
type registeredAPIListPage struct {
	RegisteredAPIs []flows.RegisteredAPI `json:"registered_apis"`
}

// RegisteredAPICmd returns the "flows registered-api" command group (Python
// globus-cli 3.42.0).
func RegisteredAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-api",
		Short: "Interact with registered APIs",
		Long:  `List and show APIs registered with Globus Flows for use as flow actions.`,
	}
	cmd.AddCommand(registeredAPIListCmd(), registeredAPIShowCmd())
	return cmd
}

func registeredAPIListCmd() *cobra.Command {
	var (
		filterRoles []string
		orderBy     []string
		limit       int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered APIs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateChoices("filter-role", filterRoles, registeredAPIRoles); err != nil {
				return err
			}
			if limit < 1 {
				return fmt.Errorf("invalid --limit %d: must be at least 1", limit)
			}
			criteria, err := parseOrderBy(orderBy, registeredAPIOrderByFields)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			opts := &flows.ListRegisteredAPIsOptions{
				FilterRoles: filterRoles,
				// Python sends one comma-joined orderby value.
				OrderBy: []string{strings.Join(criteria, ",")},
			}
			items := []flows.RegisteredAPI{}
			for {
				page, err := client.ListRegisteredAPIs(ctx, opts)
				if err != nil {
					return fmt.Errorf("failed to list registered APIs: %w", err)
				}
				items = append(items, page.RegisteredAPIs...)
				if len(items) >= limit {
					items = items[:limit]
					break
				}
				if !page.HasNextPage || page.Marker == "" {
					break
				}
				opts.Marker = page.Marker
			}
			return renderRegisteredAPIList(cmd.OutOrStdout(), items)
		},
	}
	cmd.Flags().StringArrayVar(&filterRoles, "filter-role", nil, "Filter results by the caller's role on the registered API (owner|administrator|viewer; repeatable)")
	cmd.Flags().StringArrayVar(&orderBy, "orderby", []string{"created_timestamp:DESC"}, "Sort by FIELD:ASC|DESC, FIELD one of id, name, created_timestamp, updated_timestamp (repeatable)")
	cmd.Flags().IntVar(&limit, "limit", 25, "The maximum number of results to return")
	return cmd
}

func renderRegisteredAPIList(w io.Writer, items []flows.RegisteredAPI) error {
	formatter := output.NewFormatter(viper.GetString("format"), w)
	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(registeredAPIListPage{RegisteredAPIs: items}, nil)
	}
	type row struct {
		RegisteredAPIID string
		Name            string
		CreatedAt       string
		UpdatedAt       string
	}
	rows := make([]row, 0, len(items))
	for _, a := range items {
		rows = append(rows, row{a.ID, a.Name, formatTimestamp(&a.CreatedTimestamp), formatTimestamp(&a.UpdatedTimestamp)})
	}
	return formatter.FormatOutput(rows, []string{"RegisteredAPIID", "Name", "CreatedAt", "UpdatedAt"})
}

func registeredAPIShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show REGISTERED_API_ID",
		Short: "Show a registered API",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			api, err := client.GetRegisteredAPI(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get registered API: %w", err)
			}
			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format != output.FormatText {
				return formatter.FormatOutput(api, nil)
			}
			renderRegisteredAPIRecord(cmd.OutOrStdout(), api)
			return nil
		},
	}
}

// renderRegisteredAPIRecord prints the fields of the Python CLI's
// registered_api_format_fields. Role principals are printed as URNs.
func renderRegisteredAPIRecord(w io.Writer, a *flows.RegisteredAPI) {
	fields := [][2]string{
		{"Registered API ID", a.ID},
		{"Name", a.Name},
		{"Description", a.Description},
		{"Status", a.Status},
		{"Subscription ID", a.SubscriptionID},
		{"Created At", formatTimestamp(&a.CreatedTimestamp)},
		{"Updated At", formatTimestamp(&a.UpdatedTimestamp)},
		{"Edited At", formatTimestamp(&a.EditedTimestamp)},
		{"Scheduled Deletion", formatTimestamp(a.ScheduledDeletionTimestamp)},
	}
	if a.Roles != nil {
		fields = append(fields,
			[2]string{"Owners", strings.Join(a.Roles.Owners, ", ")},
			[2]string{"Administrators", strings.Join(a.Roles.Administrators, ", ")},
			[2]string{"Viewers", strings.Join(a.Roles.Viewers, ", ")})
	}
	fields = append(fields,
		[2]string{"Target Type", nestedString(a.Target, "type")},
		[2]string{"OpenAPI Version", nestedString(a.Target, "openapi_version")},
		[2]string{"Destination Method", nestedString(a.Target, "destination", "method")},
		[2]string{"Destination URL", nestedString(a.Target, "destination", "url")})
	printRecord(w, fields)
}

// nestedString walks m along keys and returns the string found there, or "".
func nestedString(m map[string]interface{}, keys ...string) string {
	var cur interface{} = m
	for _, k := range keys {
		mm, ok := cur.(map[string]interface{})
		if !ok {
			return ""
		}
		cur = mm[k]
	}
	s, _ := cur.(string)
	return s
}
