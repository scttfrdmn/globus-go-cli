// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

package project

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

// Flag variables for the credential subcommands. Prefixed with `cred` to avoid
// collisions with other command groups in the project package.
var (
	credCreateName string
	credRotateName string
	credRotateDays int
)

// CredentialCmd returns the `credential` command tree for managing a client's
// secret credentials, including CLI-side credential rotation and scheduled
// deletion (rotation metadata is persisted in the per-profile state store,
// since the Globus Auth API has no rotation primitive).
func CredentialCmd() *cobra.Command {
	credentialCmd := &cobra.Command{
		Use:   "credential",
		Short: "Commands for managing client credentials",
		Long: `List, create, delete, and rotate the secret credentials of a Globus Auth
client, plus process scheduled credential deletions.

Rotation and scheduled deletion metadata is tracked in a per-profile state
file, since the Globus Auth API has no native rotation primitive.`,
	}

	credentialCmd.AddCommand(
		credListCmd(),
		credCreateCmd(),
		credDeleteCmd(),
		credRotateCmd(),
		credListAgeCmd(),
		credProcessDeletionsCmd(),
	)

	return credentialCmd
}

// credListCmd returns the `credential list` command.
func credListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list CLIENT_ID",
		Short: "List a client's credentials",
		Long:  `List the secret credentials of the given Globus Auth client.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return credList(cmd, args[0])
		},
	}
}

// credCreateCmd returns the `credential create` command.
func credCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create CLIENT_ID",
		Short: "Create a new client credential",
		Long: `Create a new secret credential for the given Globus Auth client.

The credential's secret is printed only once, at creation time. Store it
securely; it cannot be retrieved again.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return credCreate(cmd, args[0])
		},
	}
	cmd.Flags().StringVar(&credCreateName, "name", "", "Name for the new credential (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// credDeleteCmd returns the `credential delete` command.
func credDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete CLIENT_ID CREDENTIAL_ID",
		Short: "Delete a client credential",
		Long:  `Delete the given credential from the given Globus Auth client.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return credDelete(cmd, args[0], args[1])
		},
	}
}

// credRotateCmd returns the `credential rotate` command.
func credRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate CLIENT_ID CREDENTIAL_ID",
		Short: "Rotate a client credential",
		Long: `Rotate a client credential: create a replacement credential and schedule
the old one for deletion after a transition period.

The new credential's secret is printed only once, here. The old credential is
not deleted immediately; run "credential process-deletions" after the
transition period to delete credentials whose scheduled deletion has passed.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return credRotate(cmd, args[0], args[1])
		},
	}
	cmd.Flags().StringVar(&credRotateName, "name", "", "Name for the new credential (default \"Rotated from <old>\")")
	cmd.Flags().IntVar(&credRotateDays, "transition-days", 7, "Days until the old credential is scheduled for deletion (0 to skip scheduling)")
	return cmd
}

// credListAgeCmd returns the `credential list-age` command.
func credListAgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-age CLIENT_ID",
		Short: "List a client's credentials with their age",
		Long: `List the given client's credentials along with each credential's age in
days and any scheduled deletion time recorded in the local state store.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return credListAge(cmd, args[0])
		},
	}
}

// credProcessDeletionsCmd returns the `credential process-deletions` command.
func credProcessDeletionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "process-deletions",
		Short: "Delete credentials whose scheduled deletion has passed",
		Long: `Delete all credentials (across clients) whose scheduled deletion time,
recorded during rotation, is now in the past. A failure to delete one
credential does not stop processing of the others.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return credProcessDeletions(cmd)
		},
	}
}

// credList lists a client's credentials.
func credList(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	creds, err := client.GetClientCredentials(ctx, clientID)
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(creds, nil)
	}

	type credRow struct {
		ID      string
		Name    string
		Created string
	}
	rows := make([]credRow, 0, len(creds))
	for _, c := range creds {
		rows = append(rows, credRow{
			ID:      c.ID,
			Name:    c.Name,
			Created: c.Created.Format(time.RFC3339),
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "Name", "Created"})
}

// credCreate creates a new credential and records it in the state store.
func credCreate(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	cred, err := client.CreateClientCredential(ctx, clientID, credCreateName)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	st, err := loadState()
	if err != nil {
		return err
	}
	st.Credentials[cred.ID] = credentialRecord{ClientID: clientID, Name: credCreateName}
	if err := saveState(st); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created credential %s (%s)\n", cred.ID, cred.Name)
	if cred.Secret != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Secret: %s\n", *cred.Secret)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "The secret is shown only once; store it securely.")
	return nil
}

// credDelete deletes a credential and removes it from the state store.
func credDelete(cmd *cobra.Command, clientID, credentialID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	if err := client.DeleteClientCredential(ctx, clientID, credentialID); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	st, err := loadState()
	if err != nil {
		return err
	}
	delete(st.Credentials, credentialID)
	if err := saveState(st); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleted credential %s\n", credentialID)
	return nil
}

// credRotate rotates a credential: it creates a replacement and schedules the
// old credential for deletion, tracking the transition in the state store.
func credRotate(cmd *cobra.Command, clientID, credentialID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Verify the old credential exists.
	creds, err := client.GetClientCredentials(ctx, clientID)
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}
	found := false
	for _, c := range creds {
		if c.ID == credentialID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("credential %s not found on client %s", credentialID, clientID)
	}

	name := credRotateName
	if name == "" {
		name = fmt.Sprintf("Rotated from %s", credentialID)
	}

	newCred, err := client.CreateClientCredential(ctx, clientID, name)
	if err != nil {
		return fmt.Errorf("failed to create replacement credential: %w", err)
	}

	st, err := loadState()
	if err != nil {
		return err
	}

	now := time.Now()
	var scheduled string
	if credRotateDays > 0 {
		scheduled = now.Add(time.Duration(credRotateDays) * 24 * time.Hour).Format(time.RFC3339)
	}

	// Update (or create) the old credential's record with rotation metadata.
	oldRec := st.Credentials[credentialID]
	oldRec.ClientID = clientID
	oldRec.RotatedTo = newCred.ID
	oldRec.ScheduledDeletion = scheduled
	st.Credentials[credentialID] = oldRec

	// Ensure a record exists for the new credential.
	newRec := st.Credentials[newCred.ID]
	newRec.ClientID = clientID
	newRec.Name = name
	newRec.RotatedFrom = credentialID
	st.Credentials[newCred.ID] = newRec

	if err := saveState(st); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Rotated credential %s -> %s (%s)\n", credentialID, newCred.ID, name)
	if newCred.Secret != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Secret: %s\n", *newCred.Secret)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "The secret is shown only once; store it securely.")
	if scheduled != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Old credential %s scheduled for deletion at %s\n", credentialID, scheduled)
		fmt.Fprintln(cmd.OutOrStdout(), "Run \"globus project credential process-deletions\" after the transition period.")
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Old credential %s not scheduled for deletion (transition-days=0).\n", credentialID)
	}
	return nil
}

// credListAge lists a client's credentials with their age in days and any
// scheduled deletion recorded in the state store.
func credListAge(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	creds, err := client.GetClientCredentials(ctx, clientID)
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	st, err := loadState()
	if err != nil {
		return err
	}

	type ageRow struct {
		ID                string    `json:"id"`
		Name              string    `json:"name"`
		Created           time.Time `json:"created"`
		AgeDays           int       `json:"age_days"`
		ScheduledDeletion string    `json:"scheduled_deletion,omitempty"`
	}
	rows := make([]ageRow, 0, len(creds))
	now := time.Now()
	for _, c := range creds {
		ageDays := int(now.Sub(c.Created).Hours() / 24)
		scheduled := ""
		if rec, ok := st.Credentials[c.ID]; ok {
			scheduled = rec.ScheduledDeletion
		}
		rows = append(rows, ageRow{
			ID:                c.ID,
			Name:              c.Name,
			Created:           c.Created,
			AgeDays:           ageDays,
			ScheduledDeletion: scheduled,
		})
	}

	formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(rows, nil)
	}

	type ageTextRow struct {
		ID                string
		Name              string
		AgeDays           int
		ScheduledDeletion string
	}
	textRows := make([]ageTextRow, 0, len(rows))
	for _, r := range rows {
		textRows = append(textRows, ageTextRow{
			ID:                r.ID,
			Name:              r.Name,
			AgeDays:           r.AgeDays,
			ScheduledDeletion: r.ScheduledDeletion,
		})
	}
	return formatter.FormatOutput(textRows, []string{"ID", "Name", "AgeDays", "ScheduledDeletion"})
}

// credProcessDeletions deletes every credential whose scheduled deletion time
// has passed. One failure does not abort processing of the rest.
func credProcessDeletions(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	st, err := loadState()
	if err != nil {
		return err
	}

	now := time.Now()
	var deleted, failed int
	for id, rec := range st.Credentials {
		if rec.ScheduledDeletion == "" {
			continue
		}
		when, err := time.Parse(time.RFC3339, rec.ScheduledDeletion)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Skipping %s: invalid scheduled deletion %q: %v\n", id, rec.ScheduledDeletion, err)
			continue
		}
		if when.After(now) {
			continue
		}
		if err := client.DeleteClientCredential(ctx, rec.ClientID, id); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to delete credential %s (client %s): %v\n", id, rec.ClientID, err)
			failed++
			continue
		}
		delete(st.Credentials, id)
		deleted++
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted credential %s (client %s)\n", id, rec.ClientID)
	}

	if err := saveState(st); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Processed scheduled deletions: %d deleted, %d failed\n", deleted, failed)
	return nil
}
