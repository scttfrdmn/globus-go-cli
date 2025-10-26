// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	createFlowName      string
	createFlowInterval  string
	createFlowCron      string
	createFlowInput     string
	createFlowScope     string
	createFlowLabel     string
	createFlowStart     string
	createFlowStop      string
	createFlowTimezone  string
)

// CreateFlowCmd represents the timer create flow command
var CreateFlowCmd = &cobra.Command{
	Use:   "flow FLOW_ID",
	Short: "Create a flow timer (v3.39.0 feature)",
	Long: `Create a timer to schedule flow executions.

The timer can execute once, on a recurring schedule, or using cron syntax.

Schedule Types:
  One-time:  No interval or cron specified
  Recurring: --interval P1D (ISO 8601 duration)
  Cron:      --cron "0 0 * * *" (standard cron syntax)

ISO 8601 Duration Examples:
  P1D    - Every 1 day
  P1W    - Every 1 week
  PT1H   - Every 1 hour
  P1M    - Every 1 month

Examples:
  # One-time flow execution
  globus timer create flow FLOW_ID \
    --name "One-time Process" \
    --flow-scope FLOW_SCOPE \
    --start "2025-10-25T10:00:00Z" \
    --input '{"param1": "value1"}'

  # Daily recurring flow
  globus timer create flow FLOW_ID \
    --name "Daily Report" \
    --flow-scope FLOW_SCOPE \
    --start "2025-10-25T00:00:00Z" \
    --interval P1D \
    --input '{"report_type": "daily"}'

  # Cron-based flow (every day at midnight)
  globus timer create flow FLOW_ID \
    --name "Nightly Process" \
    --flow-scope FLOW_SCOPE \
    --start "2025-10-25T00:00:00Z" \
    --cron "0 0 * * *" \
    --timezone "America/New_York" \
    --input '{}'`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateFlowTimer,
}

func init() {
	CreateFlowCmd.Flags().StringVar(&createFlowName, "name", "", "Name for the timer (required)")
	CreateFlowCmd.Flags().StringVar(&createFlowInterval, "interval", "", "ISO 8601 interval for recurring execution (e.g., P1D, P1W, PT1H)")
	CreateFlowCmd.Flags().StringVar(&createFlowCron, "cron", "", "Cron expression for scheduled execution")
	CreateFlowCmd.Flags().StringVar(&createFlowInput, "input", "{}", "Flow input parameters as JSON string")
	CreateFlowCmd.Flags().StringVar(&createFlowScope, "flow-scope", "", "Flow scope (required for flow execution)")
	CreateFlowCmd.Flags().StringVar(&createFlowLabel, "flow-label", "", "Label for the flow run")
	CreateFlowCmd.Flags().StringVar(&createFlowStart, "start", "", "Start time (RFC3339 format, required)")
	CreateFlowCmd.Flags().StringVar(&createFlowStop, "stop", "", "Stop time (RFC3339 format)")
	CreateFlowCmd.Flags().StringVar(&createFlowTimezone, "timezone", "UTC", "Timezone for cron schedule (default: UTC)")

	CreateFlowCmd.MarkFlagRequired("name")
	CreateFlowCmd.MarkFlagRequired("start")
	CreateFlowCmd.MarkFlagRequired("flow-scope")
}

func runCreateFlowTimer(cmd *cobra.Command, args []string) error {
	flowID := args[0]

	// Validate that only one scheduling method is specified
	schedulingMethods := 0
	if createFlowInterval != "" {
		schedulingMethods++
	}
	if createFlowCron != "" {
		schedulingMethods++
	}
	if schedulingMethods > 1 {
		return fmt.Errorf("cannot specify both --interval and --cron; choose one scheduling method")
	}

	// Parse flow input JSON
	var flowInput map[string]interface{}
	if err := json.Unmarshal([]byte(createFlowInput), &flowInput); err != nil {
		return fmt.Errorf("invalid JSON in --input: %w", err)
	}

	// Parse start time (required)
	startTime, err := time.Parse(time.RFC3339, createFlowStart)
	if err != nil {
		return fmt.Errorf("invalid start time format (use RFC3339): %w", err)
	}

	// Parse stop time if provided
	var stopTime *time.Time
	if createFlowStop != "" {
		st, err := time.Parse(time.RFC3339, createFlowStop)
		if err != nil {
			return fmt.Errorf("invalid stop time format (use RFC3339): %w", err)
		}
		stopTime = &st
	}

	// Get current profile
	profile := viper.GetString("profile")

	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create authorizer
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create timers client
	timersClient, err := timers.NewClient(
		timers.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create timers client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create FlowTimer configuration
	flowTimer := &timers.FlowTimer{
		FlowID:    flowID,
		FlowScope: createFlowScope,
		FlowInput: flowInput,
		FlowLabel: createFlowLabel,
	}

	// Additional user data (empty for now)
	userData := make(map[string]interface{})

	// Create the timer using appropriate FlowTimer helper
	var createdTimer *timers.Timer
	var scheduleType string

	if createFlowCron != "" {
		// Cron-based scheduling (v3.65.0 feature)
		scheduleType = "cron"
		createdTimer, err = timersClient.CreateFlowTimerCron(
			ctx,
			createFlowName,
			createFlowCron,
			createFlowTimezone,
			stopTime,
			flowTimer,
			userData,
		)
		if err != nil {
			return fmt.Errorf("error creating cron flow timer: %w", err)
		}
	} else if createFlowInterval != "" {
		// Recurring scheduling (v3.65.0 feature)
		scheduleType = "recurring"
		createdTimer, err = timersClient.CreateFlowTimerRecurring(
			ctx,
			createFlowName,
			startTime,
			createFlowInterval,
			stopTime,
			flowTimer,
			userData,
		)
		if err != nil {
			return fmt.Errorf("error creating recurring flow timer: %w", err)
		}
	} else {
		// One-time execution (v3.65.0 feature)
		scheduleType = "once"
		createdTimer, err = timersClient.CreateFlowTimerOnce(
			ctx,
			createFlowName,
			startTime,
			flowTimer,
			userData,
		)
		if err != nil {
			return fmt.Errorf("error creating one-time flow timer: %w", err)
		}
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Flow timer created successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", createdTimer.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", createdTimer.Name)
	fmt.Fprintf(os.Stdout, "Flow ID:     %s\n", flowID)
	fmt.Fprintf(os.Stdout, "Schedule:    %s\n", scheduleType)
	if createFlowInterval != "" {
		fmt.Fprintf(os.Stdout, "Interval:    %s\n", createFlowInterval)
	}
	if createFlowCron != "" {
		fmt.Fprintf(os.Stdout, "Cron:        %s\n", createFlowCron)
		fmt.Fprintf(os.Stdout, "Timezone:    %s\n", createFlowTimezone)
	}
	if createdTimer.NextDue != nil {
		fmt.Fprintf(os.Stdout, "Next Run:    %s\n", createdTimer.NextDue.Format(time.RFC3339))
	}

	return nil
}
