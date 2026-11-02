// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 Scott Friedman and Project Contributors
package flows

import (
	"fmt"
	"strings"
	"time"
)

// parseOrderBy converts Python-CLI style "--orderby FIELD:ASC|DESC" values
// into the "FIELD ORDER" criteria the Flows API expects, validating FIELD
// against allowed. Ordering is case-insensitive and normalized to upper case.
func parseOrderBy(values []string, allowed []string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, v := range values {
		field, order, ok := strings.Cut(v, ":")
		if !ok {
			return nil, fmt.Errorf("invalid --orderby %q: expected FIELD:ASC or FIELD:DESC", v)
		}
		field = strings.ToLower(strings.TrimSpace(field))
		order = strings.ToUpper(strings.TrimSpace(order))
		if !contains(allowed, field) {
			return nil, fmt.Errorf("invalid --orderby field %q: must be one of %s", field, strings.Join(allowed, ", "))
		}
		if order != "ASC" && order != "DESC" {
			return nil, fmt.Errorf("invalid --orderby order %q: must be ASC or DESC", order)
		}
		out = append(out, field+" "+order)
	}
	return out, nil
}

// validateChoices checks that every value is in allowed, naming flag in the
// error. It mirrors click.Choice validation in the Python CLI.
func validateChoices(flag string, values []string, allowed []string) error {
	for _, v := range values {
		if !contains(allowed, v) {
			return fmt.Errorf("invalid --%s %q: must be one of %s", flag, v, strings.Join(allowed, ", "))
		}
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// formatTimestamp renders a timestamp the way the Python CLI's Date formatter
// does ("YYYY-MM-DD HH:MM:SS"), or "" for a zero/nil time.
func formatTimestamp(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}
