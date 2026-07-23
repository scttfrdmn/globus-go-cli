// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package globusauth

import "testing"

// TestDefaultLoginServicesExcludesTimers guards issue #40: the default login
// set must not request the Timers scope, whose client-specific scope a generic
// native client can't request (causing UNKNOWN_SCOPE_ERROR for the whole login).
func TestDefaultLoginServicesExcludesTimers(t *testing.T) {
	for _, svc := range DefaultLoginServices {
		if svc == ServiceTimers {
			t.Fatal("DefaultLoginServices must not include ServiceTimers (issue #40)")
		}
	}
	// Timers must still be a known service (opt-in via --scopes timers).
	if _, ok := ServiceByName("timers"); !ok {
		t.Error("ServiceTimers should remain resolvable by name for opt-in")
	}
	// Every default service must resolve to a real registry entry.
	for _, svc := range DefaultLoginServices {
		if _, ok := registry[svc]; !ok {
			t.Errorf("DefaultLoginServices contains unknown service %q", svc)
		}
	}
}

func TestServiceByName(t *testing.T) {
	tests := []struct {
		name    string
		wantOK  bool
		wantSvc Service
	}{
		{"transfer", true, ServiceTransfer},
		{"timers", true, ServiceTimers},
		{"auth", true, ServiceAuth},
		{"bogus", false, ""},
		{"", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ServiceByName(tt.name)
			if ok != tt.wantOK || got != tt.wantSvc {
				t.Errorf("ServiceByName(%q) = (%q, %v), want (%q, %v)", tt.name, got, ok, tt.wantSvc, tt.wantOK)
			}
		})
	}
}
