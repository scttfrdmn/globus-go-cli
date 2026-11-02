// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 Scott Friedman and Project Contributors
package transfer

import "testing"

func TestEndpointDomain(t *testing.T) {
	cases := []struct{ tlsftp, manager, want string }{
		{"tlsftp://abc.0ec8.data.globus.org:443", "", "abc.0ec8.data.globus.org"},
		{"tlsftp://abc.example.org:443", "https://xyz.example.org", "abc.example.org"},
		{"", "https://xyz.0ec8.data.globus.org", "xyz.0ec8.data.globus.org"},
		{"", "", "None"},
	}
	for _, c := range cases {
		if got := endpointDomain(c.tlsftp, c.manager); got != c.want {
			t.Errorf("endpointDomain(%q, %q) = %q, want %q", c.tlsftp, c.manager, got, c.want)
		}
	}
}
