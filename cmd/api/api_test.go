// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 Scott Friedman and Project Contributors
package api

import "testing"

func TestTransferPath(t *testing.T) {
	cases := map[string]string{
		"/endpoint_search":        "/v0.10/endpoint_search",
		"endpoint/abc":            "/v0.10/endpoint/abc",
		"/v0.10/task_list":        "/v0.10/task_list",
		"/v2/tunnels":             "/v2/tunnels",
		"v2/stream_access_points": "/v2/stream_access_points",
		"/v2":                     "/v2",
		"/v20/not-a-version":      "/v0.10/v20/not-a-version",
		"/v0.100/not-a-version":   "/v0.10/v0.100/not-a-version",
	}
	for in, want := range cases {
		if got := transferPath(in); got != want {
			t.Errorf("transferPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTimersAlias(t *testing.T) {
	cmd := APICmd()
	c, _, err := cmd.Find([]string{"timers"})
	if err != nil || c.Name() != "timer" {
		t.Fatalf("expected 'api timers' to resolve to 'api timer', got %v, %v", c, err)
	}
}
