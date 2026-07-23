// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package timer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/timers"
	"github.com/spf13/cobra"
)

var (
	createFlowName     string
	createFlowInterval string
	createFlowCron     string
	createFlowInput    string
	createFlowScope    string
	createFlowLabel    string
	createFlowStart    string
	createFlowStop     string
	createFlowStopRuns int
	createFlowTimezone string
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
	CreateFlowCmd.Flags().StringVar(&createFlowInterval, "interval", "", "Interval for recurring execution as a Go duration (e.g., 1h, 30m, 24h)")
	CreateFlowCmd.Flags().StringVar(&createFlowCron, "cron", "", "Deprecated: cron scheduling is not supported by the Timers API")
	CreateFlowCmd.Flags().StringVar(&createFlowInput, "input", "{}", "Flow input parameters as JSON string")
	CreateFlowCmd.Flags().StringVar(&createFlowScope, "flow-scope", "", "Flow scope (required for flow execution)")
	CreateFlowCmd.Flags().StringVar(&createFlowLabel, "flow-label", "", "Label for the flow run")
	CreateFlowCmd.Flags().StringVar(&createFlowStart, "start", "", "Start time (RFC3339 format, required)")
	CreateFlowCmd.Flags().StringVar(&createFlowStop, "stop", "", "Stop time (RFC3339 format)")
	CreateFlowCmd.Flags().IntVar(&createFlowStopRuns, "stop-after-runs", 0, "Stop running the flow after this number of runs")
	CreateFlowCmd.Flags().StringVar(&createFlowTimezone, "timezone", "UTC", "Deprecated: only used with the removed cron scheduling")

	_ = CreateFlowCmd.MarkFlagRequired("name")
	_ = CreateFlowCmd.MarkFlagRequired("start")
	_ = CreateFlowCmd.MarkFlagRequired("flow-scope")
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Timers client authorized for the current profile.
	timersClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Cron scheduling does not exist in the Globus Timers API at this SDK
	// version; only once and recurring (fixed interval) schedules are supported.
	if createFlowCron != "" {
		return fmt.Errorf("cron scheduling is not supported by the Globus Timers API; use --interval for recurring timers")
	}

	// Build the schedule (once, or recurring by fixed interval).
	var schedule *timers.Schedule
	scheduleType := "once"
	if createFlowInterval != "" {
		d, derr := time.ParseDuration(createFlowInterval)
		if derr != nil || d <= 0 {
			return fmt.Errorf("invalid --interval %q: use a Go duration such as 1h, 30m, or 24h", createFlowInterval)
		}
		if stopTime != nil && createFlowStopRuns > 0 {
			return fmt.Errorf("--stop and --stop-after-runs are mutually exclusive")
		}
		var end *timers.ScheduleEnd
		if createFlowStopRuns > 0 {
			end = &timers.ScheduleEnd{Condition: "iterations", Iterations: createFlowStopRuns}
		} else if stopTime != nil {
			end = &timers.ScheduleEnd{Condition: "time", Datetime: stopTime.Format(time.RFC3339)}
		}
		schedule = timers.NewRecurringSchedule(int(d.Seconds()), startTime.Format(time.RFC3339), end)
		scheduleType = "recurring"
	} else {
		schedule = timers.NewOnceSchedule(startTime.Format(time.RFC3339))
	}

	// Build the flow timer document. Flow input is carried in the body; scope
	// and label are supplied through the body when the flow requires them.
	body := map[string]interface{}{"body": flowInput}
	if createFlowLabel != "" {
		body["label"] = createFlowLabel
	}
	if createFlowScope != "" {
		body["scope"] = createFlowScope
	}
	flowTimer := timers.NewFlowTimer(createFlowName, flowID, schedule, body)

	createdTimer, err := timersClient.CreateTimer(ctx, flowTimer)
	if err != nil {
		return fmt.Errorf("error creating flow timer: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Flow timer created successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Timer ID:    %s\n", createdTimer.JobID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", createdTimer.Name)
	fmt.Fprintf(os.Stdout, "Flow ID:     %s\n", flowID)
	fmt.Fprintf(os.Stdout, "Schedule:    %s\n", scheduleType)
	if createFlowInterval != "" {
		fmt.Fprintf(os.Stdout, "Interval:    %s\n", createFlowInterval)
	}
	if !createdTimer.NextRun.IsZero() {
		fmt.Fprintf(os.Stdout, "Next Run:    %s\n", createdTimer.NextRun.Format(time.RFC3339))
	}

	return nil
}
