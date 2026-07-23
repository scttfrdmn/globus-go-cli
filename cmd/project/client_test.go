// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package project

import (
	"errors"
	"reflect"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// TestSessionRequiredPolicies covers extraction of
// authorization_parameters.session_required_policies from a Globus 403 (issue
// #41), across the shapes the field can take on the wire.
func TestSessionRequiredPolicies(t *testing.T) {
	policyErr := func(sp interface{}) error {
		return &core.APIError{
			StatusCode: 403,
			Details: map[string]interface{}{
				"authorization_parameters": map[string]interface{}{
					"session_required_policies": sp,
				},
			},
		}
	}

	tests := []struct {
		name string
		err  error
		want []string
	}{
		{
			name: "array of strings",
			err:  policyErr([]interface{}{"pol-1", "pol-2"}),
			want: []string{"pol-1", "pol-2"},
		},
		{
			name: "comma-joined string",
			err:  policyErr("pol-1"),
			want: []string{"pol-1"},
		},
		{
			name: "not a 403",
			err:  &core.APIError{StatusCode: 401, Details: map[string]interface{}{}},
			want: nil,
		},
		{
			name: "403 without authorization_parameters",
			err:  &core.APIError{StatusCode: 403, Details: map[string]interface{}{"code": "FORBIDDEN"}},
			want: nil,
		},
		{
			name: "non-api error",
			err:  errors.New("boom"),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sessionRequiredPolicies(tt.err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sessionRequiredPolicies = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSessionRequiredPoliciesWrapped verifies errors.As unwrapping finds a
// wrapped *core.APIError carrying the policies.
func TestSessionRequiredPoliciesWrapped(t *testing.T) {
	inner := &core.APIError{
		StatusCode: 403,
		Details: map[string]interface{}{
			"authorization_parameters": map[string]interface{}{
				"session_required_policies": []interface{}{"pol-x"},
			},
		},
	}
	wrapped := fmtWrap(inner)
	got := sessionRequiredPolicies(wrapped)
	if !reflect.DeepEqual(got, []string{"pol-x"}) {
		t.Errorf("sessionRequiredPolicies(wrapped) = %v, want [pol-x]", got)
	}
}

// fmtWrap wraps err so it is still discoverable via errors.As.
func fmtWrap(err error) error {
	return &wrapErr{err}
}

type wrapErr struct{ err error }

func (w *wrapErr) Error() string { return "wrapped: " + w.err.Error() }
func (w *wrapErr) Unwrap() error { return w.err }

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want []string
	}{
		{"nil", nil, nil},
		{"empty string", "", nil},
		{"single string", "a", []string{"a"}},
		{"string slice", []string{"a", "b"}, []string{"a", "b"}},
		{"interface slice with empties", []interface{}{"a", "", "b"}, []string{"a", "b"}},
		{"interface slice non-strings dropped", []interface{}{"a", 5, "b"}, []string{"a", "b"}},
		{"unexpected type", 42, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSlice(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toStringSlice(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
