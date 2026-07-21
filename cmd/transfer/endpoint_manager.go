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
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// EndpointManagerCmd returns the "endpoint-manager" command group, wrapping the
// SDK's EndpointManager* administrative family. It mirrors the Python CLI's
// "globus endpoint-manager ..." command tree for activity managers and monitors.
func EndpointManagerCmd() *cobra.Command {
	emCmd := &cobra.Command{
		Use:   "endpoint-manager",
		Short: "Administrative commands for managed endpoints",
		Long: `Administrative (activity manager/monitor) commands for endpoints
covered by a managed endpoint subscription.

These commands operate on the Globus Transfer endpoint-manager API and require
an appropriate management role on the target endpoints.`,
	}

	emCmd.AddCommand(
		endpointManagerMonitoredEndpointsCmd(),
		endpointManagerHostedEndpointListCmd(),
		endpointManagerShowCmd(),
		endpointManagerACLListCmd(),
		endpointManagerTaskListCmd(),
		endpointManagerTaskShowCmd(),
		endpointManagerTaskEventListCmd(),
		endpointManagerTaskPauseInfoCmd(),
		endpointManagerTaskCancelCmd(),
		endpointManagerCancelStatusCmd(),
		endpointManagerTaskPauseCmd(),
		endpointManagerTaskResumeCmd(),
		endpointManagerPauseRuleCmd(),
	)

	return emCmd
}

func endpointManagerMonitoredEndpointsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "monitored-endpoints",
		Short: "List endpoints the current user can monitor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerMonitoredEndpoints(ctx)
			if err != nil {
				return fmt.Errorf("failed to list monitored endpoints: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerHostedEndpointListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hosted-endpoint-list ENDPOINT_ID",
		Short: "List endpoints hosted on a managed endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerHostedEndpointList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list hosted endpoints: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID",
		Short: "Show a managed endpoint as an administrator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerGetEndpoint(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get endpoint: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerACLListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "acl-list ENDPOINT_ID",
		Short: "List access rules on a managed endpoint as an administrator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerACLList(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list endpoint access rules: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerTaskListCmd() *cobra.Command {
	var (
		filterStatus   []string
		filterEndpoint string
		filterOwnerID  string
		taskLimit      int
	)

	cmd := &cobra.Command{
		Use:   "task-list",
		Short: "List tasks on managed endpoints as an administrator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			options := &transfer.EndpointManagerTaskListOptions{
				FilterStatus:   filterStatus,
				FilterOwnerID:  filterOwnerID,
				FilterEndpoint: filterEndpoint,
				Limit:          taskLimit,
			}

			resp, err := client.EndpointManagerTaskList(ctx, options)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}
			return formatTypedData(cmd, resp.Data)
		},
	}

	cmd.Flags().StringSliceVar(&filterStatus, "filter-status", nil, "Filter by task status (repeatable or comma-separated)")
	cmd.Flags().StringVar(&filterEndpoint, "filter-endpoint", "", "Filter by endpoint ID")
	cmd.Flags().StringVar(&filterOwnerID, "filter-owner-id", "", "Filter by task owner identity ID")
	cmd.Flags().IntVar(&taskLimit, "limit", 25, "Maximum number of tasks to return")

	return cmd
}

func endpointManagerTaskShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task-show TASK_ID",
		Short: "Show a task as an administrator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerGetTask(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerTaskEventListCmd() *cobra.Command {
	var eventLimit int

	cmd := &cobra.Command{
		Use:   "task-event-list TASK_ID",
		Short: "List events for a task as an administrator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			options := &transfer.ListTaskEventsOptions{Limit: eventLimit}

			resp, err := client.EndpointManagerTaskEventList(ctx, args[0], options)
			if err != nil {
				return fmt.Errorf("failed to list task events: %w", err)
			}
			return formatTypedData(cmd, resp.Data)
		},
	}

	cmd.Flags().IntVar(&eventLimit, "limit", 25, "Maximum number of events to return")

	return cmd
}

func endpointManagerTaskPauseInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task-pause-info TASK_ID",
		Short: "Show pause information for a task as an administrator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerTaskPauseInfo(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get task pause info: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerTaskCancelCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "task-cancel TASK_ID...",
		Short: "Cancel one or more tasks as an administrator",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerCancelTasks(ctx, args, message)
			if err != nil {
				return fmt.Errorf("failed to cancel tasks: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "Message describing the reason for cancellation")

	return cmd
}

func endpointManagerCancelStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-status ADMIN_CANCEL_ID",
		Short: "Show the status of an administrative cancel request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerCancelStatus(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get cancel status: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerTaskPauseCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "task-pause TASK_ID...",
		Short: "Pause one or more tasks as an administrator",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerPauseTasks(ctx, args, message)
			if err != nil {
				return fmt.Errorf("failed to pause tasks: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "Message describing the reason for pausing")

	return cmd
}

func endpointManagerTaskResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "task-resume TASK_ID...",
		Short: "Resume one or more tasks as an administrator",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerResumeTasks(ctx, args)
			if err != nil {
				return fmt.Errorf("failed to resume tasks: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

// endpointManagerPauseRuleCmd returns the "endpoint-manager pause-rule" command
// group for managing administrative pause rules.
func endpointManagerPauseRuleCmd() *cobra.Command {
	pauseRuleCmd := &cobra.Command{
		Use:   "pause-rule",
		Short: "Manage administrative pause rules",
		Long:  `List, show, create, and delete administrative pause rules on managed endpoints.`,
	}

	pauseRuleCmd.AddCommand(
		endpointManagerPauseRuleListCmd(),
		endpointManagerPauseRuleShowCmd(),
		endpointManagerPauseRuleCreateCmd(),
		endpointManagerPauseRuleDeleteCmd(),
	)

	return pauseRuleCmd
}

func endpointManagerPauseRuleListCmd() *cobra.Command {
	var filterEndpoint string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List administrative pause rules",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerPauseRuleList(ctx, filterEndpoint)
			if err != nil {
				return fmt.Errorf("failed to list pause rules: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&filterEndpoint, "filter-endpoint", "", "Filter by endpoint ID")

	return cmd
}

func endpointManagerPauseRuleShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show PAUSE_RULE_ID",
		Short: "Show an administrative pause rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerGetPauseRule(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get pause rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}
}

func endpointManagerPauseRuleCreateCmd() *cobra.Command {
	var (
		endpointID string
		message    string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an administrative pause rule",
		Long: `Create an administrative pause rule on a managed endpoint.

This builds a minimal pause_rule document from the --endpoint and --message
flags. More granular pause rule fields are not currently exposed.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			doc := map[string]interface{}{
				"DATA_TYPE":   "pause_rule",
				"endpoint_id": endpointID,
				"message":     message,
			}

			resp, err := client.EndpointManagerCreatePauseRule(ctx, doc)
			if err != nil {
				return fmt.Errorf("failed to create pause rule: %w", err)
			}
			return formatGenericResponse(cmd, resp)
		},
	}

	cmd.Flags().StringVar(&endpointID, "endpoint", "", "Endpoint ID the pause rule applies to")
	cmd.Flags().StringVar(&message, "message", "", "Message shown to affected users")
	_ = cmd.MarkFlagRequired("endpoint")

	return cmd
}

func endpointManagerPauseRuleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete PAUSE_RULE_ID",
		Short: "Delete an administrative pause rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.EndpointManagerDeletePauseRule(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to delete pause rule: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pause rule %s deleted.\n", args[0])
			printResponseCodeMessage(cmd, resp)
			return nil
		},
	}
}

// formatTypedData routes the .Data slice of a typed list response through the
// shared formatter so -F (text/json/unix) and --jmespath/--jq work uniformly.
func formatTypedData(cmd *cobra.Command, data []map[string]interface{}) error {
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.FormatOutput(data, nil)
}
