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

// Choices accepted by "flows web-input list", matching Python globus-cli 3.43.0.
var (
	webInputRoles         = []string{"viewer", "respondent"}
	webInputStates        = []string{"open", "closed"}
	webInputOrderByFields = []string{"created_timestamp", "edited_timestamp", "closed_timestamp"}
)

// webInputListPage is the JSON document emitted by "web-input list -F json",
// matching the Python CLI's json_conversion_key ("web_input_summaries").
type webInputListPage struct {
	WebInputSummaries []flows.WebInputSummary `json:"web_input_summaries"`
}

// WebInputCmd returns the "flows web-input" command group (Python globus-cli
// 3.43.0): list, show and respond to the human-in-the-loop prompts a running
// flow raises.
func WebInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web-input",
		Short: "Interact with flow web inputs",
		Long: `Interact with web inputs: prompts a running flow raises for a person to
answer before the run can continue.`,
	}
	cmd.AddCommand(webInputListCmd(), webInputShowCmd(), webInputRespondCmd())
	return cmd
}

func webInputListCmd() *cobra.Command {
	var (
		filterRoles   []string
		filterStates  []string
		filterFlowIDs []string
		filterRunIDs  []string
		orderBy       []string
		limit         int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List web inputs accessible to the current user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateChoices("filter-role", filterRoles, webInputRoles); err != nil {
				return err
			}
			if err := validateChoices("filter-state", filterStates, webInputStates); err != nil {
				return err
			}
			if limit < 1 {
				return fmt.Errorf("invalid --limit %d: must be at least 1", limit)
			}
			criteria, err := parseOrderBy(orderBy, webInputOrderByFields)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			opts := &flows.ListWebInputsOptions{
				FilterRoles:   filterRoles,
				FilterStates:  filterStates,
				FilterFlowIDs: filterFlowIDs,
				FilterRunIDs:  filterRunIDs,
				// Python sends one comma-joined orderby value.
				OrderBy: []string{strings.Join(criteria, ",")},
			}
			items, err := collectWebInputs(ctx, client, opts, limit)
			if err != nil {
				return fmt.Errorf("failed to list web inputs: %w", err)
			}
			return renderWebInputList(cmd.OutOrStdout(), items)
		},
	}
	cmd.Flags().StringArrayVar(&filterRoles, "filter-role", nil, "Filter results by the caller's role on the web input (viewer|respondent; repeatable)")
	cmd.Flags().StringArrayVar(&filterStates, "filter-state", nil, "Filter results by the web input's state (open|closed; repeatable)")
	cmd.Flags().StringArrayVar(&filterFlowIDs, "filter-flow-id", nil, "Filter results to web inputs of this flow ID (repeatable)")
	cmd.Flags().StringArrayVar(&filterRunIDs, "filter-run-id", nil, "Filter results to web inputs of this run ID (repeatable)")
	cmd.Flags().StringArrayVar(&orderBy, "orderby", []string{"created_timestamp:DESC"}, "Sort by FIELD:ASC|DESC, FIELD one of created_timestamp, edited_timestamp, closed_timestamp (repeatable)")
	cmd.Flags().IntVar(&limit, "limit", 25, "The maximum number of results to return")
	return cmd
}

// collectWebInputs follows marker pagination until limit items are gathered
// or the service reports no further pages.
func collectWebInputs(ctx context.Context, client *flows.Client, opts *flows.ListWebInputsOptions, limit int) ([]flows.WebInputSummary, error) {
	items := []flows.WebInputSummary{}
	for {
		page, err := client.ListWebInputs(ctx, opts)
		if err != nil {
			return nil, err
		}
		items = append(items, page.WebInputSummaries...)
		if len(items) >= limit {
			return items[:limit], nil
		}
		if !page.HasNextPage() {
			return items, nil
		}
		opts.Marker = page.Marker
	}
}

func renderWebInputList(w io.Writer, items []flows.WebInputSummary) error {
	formatter := output.NewFormatter(viper.GetString("format"), w)
	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(webInputListPage{WebInputSummaries: items}, nil)
	}
	type row struct {
		WebInputID string
		Status     string
		InputType  string
		Title      string
		FlowTitle  string
		RunLabel   string
	}
	rows := make([]row, 0, len(items))
	for _, wi := range items {
		rows = append(rows, row{wi.ID, wi.Status, wi.InputType, wi.Title, wi.Flow.Title, wi.Run.Label})
	}
	return formatter.FormatOutput(rows, []string{"WebInputID", "Status", "InputType", "Title", "FlowTitle", "RunLabel"})
}

func webInputShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show WEB_INPUT_ID",
		Short: "Show a web input",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			wi, err := client.GetWebInput(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get web input: %w", err)
			}
			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format != output.FormatText {
				return formatter.FormatOutput(wi, nil)
			}
			return renderWebInputRecord(cmd.OutOrStdout(), wi)
		},
	}
}

// renderWebInputRecord prints the record fields of the Python CLI's
// web_input_format_fields. Principal URNs are printed as-is (the Python CLI
// resolves them to usernames via Auth).
func renderWebInputRecord(w io.Writer, wi *flows.WebInput) error {
	title, _ := wi.Context["title"].(string)
	fields := [][2]string{
		{"Web Input ID", wi.ID},
		{"Status", wi.Status},
		{"Title", title},
		{"Your Roles", strings.Join(wi.UserRoles, ", ")},
		{"Flow", labelledID(wi.Flow.Title, wi.Flow.ID)},
		{"Run", labelledID(wi.Run.Label, wi.Run.ID)},
		{"Creator", wi.CreatorURN},
	}
	if wi.Roles != nil {
		fields = append(fields,
			[2]string{"Viewers", strings.Join(wi.Roles.ViewerURNs, ", ")},
			[2]string{"Respondents", strings.Join(wi.Roles.RespondentURNs, ", ")})
	}
	fields = append(fields,
		[2]string{"Created At", formatTimestamp(&wi.CreatedTimestamp)},
		[2]string{"Closed At", formatTimestamp(wi.ClosedTimestamp)})
	printRecord(w, fields)

	if len(wi.Options) > 0 {
		fmt.Fprintln(w, "\nOptions:")
		for _, o := range wi.Options {
			line := "  " + o.OptionID
			if o.Label != "" {
				line += "  " + o.Label
			}
			if o.Description != "" {
				line += " — " + o.Description
			}
			fmt.Fprintln(w, line)
		}
	}
	return nil
}

func webInputRespondCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "respond WEB_INPUT_ID OPTION_ID",
		Short: "Respond to a web input",
		Long: `Respond to a web input.

For "selection" type web inputs, OPTION_ID is the ID of one of the web input's
available options (see "globus flows web-input show").`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			client, err := getClient(ctx)
			if err != nil {
				return err
			}
			resp, err := client.RespondToWebInput(ctx, args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to respond to web input: %w", err)
			}
			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format != output.FormatText {
				return formatter.FormatOutput(resp, nil)
			}
			status, _ := resp["status"].(string)
			printRecord(cmd.OutOrStdout(), [][2]string{{"Status", status}})
			return nil
		},
	}
}

// labelledID renders "label (id)", or just the id when there is no label.
func labelledID(label, id string) string {
	if label == "" {
		return id
	}
	return fmt.Sprintf("%s (%s)", label, id)
}

// printRecord prints key/value pairs as an aligned record, like the Python
// CLI's RECORD text mode.
func printRecord(w io.Writer, fields [][2]string) {
	width := 0
	for _, f := range fields {
		if len(f[0]) > width {
			width = len(f[0])
		}
	}
	for _, f := range fields {
		fmt.Fprintf(w, "%-*s %s\n", width+1, f[0]+":", f[1])
	}
}
